package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword превращает обычный пароль в bcrypt-хеш.
//
// В базу данных мы сохраняем только результат этой функции,
// но никогда не сам пароль.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword понадобится нам на следующем этапе,
// когда будем делать авторизацию.
func CheckPassword(passwordHash string, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)

	return err == nil
}
