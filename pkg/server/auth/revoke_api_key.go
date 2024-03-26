package auth

import "github.com/daytonaio/daytona/pkg/server/db"

func RevokeApiKey(name string) error {
	apiKey, err := db.FindApiKeyByName(name)
	if err != nil {
		return err
	}

	return db.DeleteApiKey(apiKey.KeyHash)
}
