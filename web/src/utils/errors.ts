const TRANSLATIONS: Record<string, string> = {
  'invalid password':              'Senha incorreta.',
  'URL not found or expired':      'Este link não existe ou já expirou.',
  'invalid manage token':          'Token de gerenciamento inválido.',
  'slug is taken and no alternative could be found':
    'Este slug já está em uso e não foi possível encontrar uma alternativa disponível.',
  'slug must be 5-50 characters: letters, numbers, or hyphens':
    'O slug deve ter entre 5 e 50 caracteres: apenas letras, números e hífens.',
  'invalid ttl value':             'Valor de expiração inválido.',
  'failed to create URL':          'Falha ao criar a URL. Tente novamente.',
  'failed to check slug':          'Falha ao verificar a disponibilidade do slug.',
  'internal error':                'Erro interno. Tente novamente mais tarde.',
  'Too Many Requests':             'Muitas requisições. Aguarde um momento e tente novamente.',
}

export function translateError(message: string): string {
  if (message.startsWith('target_url:')) {
    return 'A URL de destino deve ser um endereço http ou https válido.'
  }
  return TRANSLATIONS[message] ?? message
}
