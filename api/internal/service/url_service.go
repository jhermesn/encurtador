package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"encurtador/internal/model"
	"encurtador/internal/repository"
)

const (
	autoSlugLength    = 8
	manageTokenLength = 32
	maxCollisionTries = 10
	maxAutoSlugTries  = 10
	slugMinLength     = 5
	slugMaxLength     = 50
	base62Chars       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9-]{` + strconv.Itoa(slugMinLength) + `,` + strconv.Itoa(slugMaxLength) + `}$`)

var (
	ErrSlugTaken          = errors.New("slug is taken and no alternative could be found")
	ErrInvalidSlugFormat  = errors.New("slug must be " + strconv.Itoa(slugMinLength) + "-" + strconv.Itoa(slugMaxLength) + " characters: letters, numbers, or hyphens")
	ErrInvalidTTL         = errors.New("invalid TTL value")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidManageToken = errors.New("invalid manage token")
)

type CreateRequest struct {
	TargetURL string
	Slug      string
	TTL       model.TTL
	Password  string
}

type CreateResult struct {
	Slug        string
	ShortURL    string
	ExpiresAt   time.Time
	Protected   bool
	ManageToken string
}

type URLService struct {
	repo    repository.URLRepository
	cache   repository.URLCache
	baseURL string
}

func NewURLService(repo repository.URLRepository, cache repository.URLCache, baseURL string) *URLService {
	return &URLService{repo: repo, cache: cache, baseURL: baseURL}
}

func (s *URLService) Create(ctx context.Context, req CreateRequest) (*CreateResult, error) {
	ttlDuration, ok := model.ValidTTLs[req.TTL]
	if !ok {
		return nil, ErrInvalidTTL
	}

	slug, err := s.resolveSlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}

	var passwordHash *string
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hashing password: %w", err)
		}
		h := string(hash)
		passwordHash = &h
	}

	manageToken, manageTokenHash, err := generateManageToken()
	if err != nil {
		return nil, fmt.Errorf("generating manage token: %w", err)
	}

	expiresAt := time.Now().Add(ttlDuration)
	url := &model.URL{
		Slug:            slug,
		TargetURL:       req.TargetURL,
		PasswordHash:    passwordHash,
		ManageTokenHash: manageTokenHash,
		ExpiresAt:       expiresAt,
	}

	if err := s.repo.Create(ctx, url); err != nil {
		return nil, err
	}

	// Cache write failure is non-fatal: the redirect path will fall back to MySQL.
	if err := s.cache.Set(ctx, slug, url.ToCached(), ttlDuration); err != nil {
		slog.Warn("failed to pre-warm cache", "slug", slug, "error", err)
	}

	return &CreateResult{
		Slug:        slug,
		ShortURL:    s.baseURL + "/" + slug,
		ExpiresAt:   expiresAt,
		Protected:   passwordHash != nil,
		ManageToken: manageToken,
	}, nil
}

func (s *URLService) Resolve(ctx context.Context, slug string) (*model.CachedURL, error) {
	return s.lookupCached(ctx, slug)
}

func (s *URLService) VerifyPassword(ctx context.Context, slug, password string) (string, error) {
	cached, err := s.lookupCached(ctx, slug)
	if err != nil {
		return "", err
	}
	if cached == nil {
		return "", nil
	}
	if !cached.Protected {
		return cached.TargetURL, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cached.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidPassword
	}
	return cached.TargetURL, nil
}

// lookupCached implements the cache-aside pattern: it tries Redis first, then
// falls back to MySQL and repopulates the cache on a miss. Returns nil without
// an error when the slug does not exist or has expired.
func (s *URLService) lookupCached(ctx context.Context, slug string) (*model.CachedURL, error) {
	cached, err := s.cache.Get(ctx, slug)
	if err != nil {
		slog.Warn("cache get failed, falling back to db", "slug", slug, "error", err)
	}
	if cached != nil {
		return cached, nil
	}

	url, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, nil
	}

	remaining := time.Until(url.ExpiresAt)
	if remaining <= 0 {
		return nil, nil
	}

	cached = url.ToCached()
	if err := s.cache.Set(ctx, slug, cached, remaining); err != nil {
		slog.Warn("failed to populate cache", "slug", slug, "error", err)
	}

	return cached, nil
}

func (s *URLService) ExpireEarly(ctx context.Context, slug, manageToken string) error {
	sum := sha256.Sum256([]byte(manageToken))
	hash := hex.EncodeToString(sum[:])

	updated, err := s.repo.ExpireBySlug(ctx, slug, hash)
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidManageToken
	}

	if err := s.cache.Delete(ctx, slug); err != nil {
		slog.Warn("failed to invalidate cache after early expire", "slug", slug, "error", err)
	}
	return nil
}

func (s *URLService) CheckSlug(ctx context.Context, slug string) (available bool, suggestion string, err error) {
	if !slugPattern.MatchString(slug) {
		return false, "", ErrInvalidSlugFormat
	}

	exists, err := s.repo.SlugExists(ctx, slug)
	if err != nil {
		return false, "", err
	}
	if !exists {
		return true, "", nil
	}

	suggestion, err = s.suggestAlternative(ctx, slug)
	if err != nil {
		return false, "", err
	}
	return false, suggestion, nil
}

// RunCleanup periodically removes expired URL records. Intended to run as a goroutine.
func (s *URLService) RunCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.repo.DeleteExpired(ctx); err != nil {
				slog.Error("periodic cleanup failed", "error", err)
			}
		}
	}
}

func (s *URLService) resolveSlug(ctx context.Context, requested string) (string, error) {
	if requested == "" {
		return s.generateUniqueSlug(ctx)
	}

	if !slugPattern.MatchString(requested) {
		return "", ErrInvalidSlugFormat
	}

	exists, err := s.repo.SlugExists(ctx, requested)
	if err != nil {
		return "", err
	}
	if !exists {
		return requested, nil
	}

	candidate, err := s.suggestAlternative(ctx, requested)
	if err != nil {
		return "", err
	}
	if candidate != "" {
		return candidate, nil
	}
	return "", ErrSlugTaken
}

func (s *URLService) generateUniqueSlug(ctx context.Context) (string, error) {
	for range maxAutoSlugTries {
		slug, err := randomBase62(autoSlugLength)
		if err != nil {
			return "", fmt.Errorf("generating random slug: %w", err)
		}

		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique slug after %d attempts", maxAutoSlugTries)
}

// suggestAlternative finds the first available "slug-N" variant, starting at N=2.
// Returns an empty string (without error) if all candidates are taken.
func (s *URLService) suggestAlternative(ctx context.Context, slug string) (string, error) {
	for i := 2; i <= maxCollisionTries; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		exists, err := s.repo.SlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", nil
}

// randomBase62 generates a cryptographically random base62 string of the given
// length. Rejection sampling is used to eliminate modulo bias: bytes >= 248
// (i.e. 256 - 256%62) are discarded so every character has equal probability.
func randomBase62(length int) (string, error) {
	const unbiasedCeiling = 256 - 256%len(base62Chars) // 248
	result := make([]byte, 0, length)
	buf := make([]byte, length)
	for len(result) < length {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) < unbiasedCeiling {
				result = append(result, base62Chars[int(b)%len(base62Chars)])
				if len(result) == length {
					break
				}
			}
		}
	}
	return string(result), nil
}

func generateManageToken() (string, string, error) {
	plain, err := randomBase62(manageTokenLength)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(sum[:]), nil
}
