// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

func (s *ApiKeyService) Revoke(name string) error {
	apiKey, err := s.apiKeyStore.FindByName(name)
	if err != nil {
		return err
	}

	return s.apiKeyStore.Delete(apiKey)
}
