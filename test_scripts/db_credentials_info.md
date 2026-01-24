# 데이터베이스 등록 Credential 정보

## Docker Registry Credentials

데이터베이스에 등록된 모든 Docker Registry credential 정보입니다.

### 조회 결과

| ID | Registry URL | Username | Password Length |
|----|--------------|----------|-----------------|
| 00000000-0000-0000-0000-000000000001 | index.docker.io/v1/ | hyoungjunnoh | 36자 |
| (3개 추가 레지스트리) | - | - | - |

### 상세 정보

#### 1. Docker Hub (공식)
- **ID**: `00000000-0000-0000-0000-000000000001`
- **URL**: `index.docker.io/v1/`
- **Username**: `hyoungjunnoh`
- **Password**: `dckr_p...` (36자, Docker Personal Access Token)
- **용도**: Docker Hub 공식 이미지 pull 시 사용

#### 추가 레지스트리
데이터베이스에는 총 4개의 레지스트리가 등록되어 있습니다. 나머지 3개는 내부 레지스트리이거나 다른 용도로 사용되는 것으로 보입니다.

### 확인 방법

전체 credential 목록 조회:
```bash
docker exec daytona-db-1 psql -U user -d daytona -c "SELECT id, url, username, LENGTH(password) as pwd_len FROM docker_registry;"
```

특정 레지스트리 조회:
```bash
docker exec daytona-db-1 psql -U user -d daytona -c "SELECT * FROM docker_registry WHERE url LIKE '%docker%';"
```
