# 커밋 검토 및 정리 계획

## 1. 수정된 파일 (Modified Files)

### ✅ 커밋 필요: apps/runner/Dockerfile.local
**변경 내용**:
- Line 14: `ENV LOG_LEVEL=info` 추가 (credential 로깅 가시성)
- Line 36: ENTRYPOINT 수정 (Docker daemon 준비 대기)

**커밋 이유**:
- Runner stability fix의 핵심 부분
- Docker daemon race condition 해결
- Credential logging 활성화

**커밋 메시지**: 
```
fix(runner): Wait for Docker daemon and enable credential logging

- Add while loop to wait for /var/run/docker.sock before starting runner
- Set LOG_LEVEL=info to make credential usage visible in logs
- Fixes race condition where runner starts before dockerd is ready

Resolves: Runner "Is the docker daemon running?" errors
```

### ⚠️ 로컬 전용: docker/docker-compose.yaml
**변경 내용 검토 필요**:
- pgadmin 설정 변경
- 기타 로컬 개발 환경 설정

**제안**: 
- 파일명을 `docker-compose.local.yaml`로 복사
- 원본 `docker-compose.yaml`은 git restore
- `.gitignore`에 `docker-compose.local.yaml` 추가

### ❌ 커밋 불필요: .dockerignore
**이유**: 로컬 개발 환경 설정

---

## 2. Python SDK 관련 파일 (Untracked - API Changes)

### ✅ 커밋 필요: apps/api/src/organization/controllers/default-region.controller.ts
**이유**: Python SDK `/api/region` 엔드포인트 지원 (이미 커밋됨)

### ✅ 커밋 필요: apps/api/src/organization/organization.module.ts
**이유**: DefaultRegionController 등록 (이미 커밋됨)

### ✅ 커밋 필요: apps/api/src/organization/controllers/region.controller.ts
**이유**: includeShared 기본값 true 설정 (이미 커밋됨)

---

## 3. Scripts 디렉토리 정리

### ✅ 커밋 권장 (중요 테스트/검증 스크립트)

#### scripts/test_sdk.py
- **용도**: Python SDK 통합 테스트
- **중요도**: ⭐⭐⭐⭐⭐
- **이유**: SDK 사용 예제 및 검증 스크립트

#### scripts/verify_credential_logging.py
- **용도**: Docker credential 로깅 검증
- **중요도**: ⭐⭐⭐⭐
- **이유**: Credential passing 검증 자동화

#### scripts/test_dockerhub_fix.py
- **용도**: Docker Hub 인증 수정 검증
- **중요도**: ⭐⭐⭐⭐
- **이유**: Bearer token 인증 테스트

#### scripts/requirements.txt
- **용도**: 테스트 스크립트 의존성
- **중요도**: ⭐⭐⭐
- **이유**: 스크립트 실행에 필요

### ⚠️ 선택적 커밋 (참고용)

#### scripts/troubleshooting_log.md
- **용도**: 디버깅 과정 기록
- **중요도**: ⭐⭐⭐
- **이유**: 문제 해결 과정 문서화 (이미 troubleshooting_report_pr.md에 포함)

### ❌ 커밋 불필요 (로컬 전용/임시 파일)

- `scripts/.env` - 로컬 환경 변수
- `scripts/check_status.py` - 임시 디버깅 스크립트
- `scripts/db_credentials_info.md` - 로컬 DB 정보
- `scripts/show_db_credentials.py` - 임시 스크립트
- `scripts/simple_creds.py` - 임시 스크립트
- `scripts/repro_sdk_issue.py` - 디버깅용
- `scripts/test_sdk_trace.py` - 디버깅용
- `scripts/test_snapshot_creation.py` - 임시 테스트
- `scripts/test_sandbox_creation.py` - 임시 테스트
- `scripts/test_output.log` - 로그 파일
- `scripts/onboarding.py` - 기존 예제 (수정 없음)
- `scripts/setup-proxy-dns.ps1` - 로컬 환경 설정
- `scripts/test_create_sandbox.ps1` - 로컬 테스트

---

## 4. 루트 디렉토리 파일 정리

### ❌ 커밋 불필요 (모두 로컬/임시 파일)

**SQL 파일들** (로컬 DB 조작용):
- `assign_region.sql`
- `check_postgres_error.sql`
- `check_runner.sql`
- `fix.sql`
- `fix_docker_hub.sql`
- `fix_runner.sql`
- `update_real_creds.sql`

**임시 데이터 파일들**:
- `all_registries.txt`
- `all_snapshots.txt`
- `last_snapshot.txt`
- `postgres_error.txt`
- `registries.txt`
- `runner_snapshots.txt`
- `runner_snapshots_refs.txt`
- `sandboxes.json`
- `sdk_trace.log`
- `verify.log`

**Python 스크립트들** (임시/디버깅):
- `daytona_create.py`
- `daytona_full.py`

---

## 5. 커밋 실행 계획

### Phase 1: Runner Dockerfile 커밋
```bash
git add apps/runner/Dockerfile.local
git commit -m "fix(runner): Wait for Docker daemon and enable credential logging"
```

### Phase 2: 중요 테스트 스크립트 커밋
```bash
git add scripts/test_sdk.py scripts/verify_credential_logging.py scripts/test_dockerhub_fix.py scripts/requirements.txt
git commit -m "test: Add verification scripts for credential passing and SDK integration"
```

### Phase 3: docker-compose.yaml 처리
```bash
# 로컬 버전 백업
cp docker/docker-compose.yaml docker/docker-compose.local.yaml

# 원본 복원
git restore docker/docker-compose.yaml

# .gitignore 업데이트
echo "docker/docker-compose.local.yaml" >> .gitignore
```

### Phase 4: 불필요한 파일 정리
```bash
# .gitignore에 추가
echo "*.log" >> .gitignore
echo "*.txt" >> .gitignore (선택적)
echo "*.sql" >> .gitignore (루트의 임시 SQL만)
```

---

## 6. 최종 권장사항

### 커밋할 파일:
1. ✅ `apps/runner/Dockerfile.local`
2. ✅ `scripts/test_sdk.py`
3. ✅ `scripts/verify_credential_logging.py`
4. ✅ `scripts/test_dockerhub_fix.py`
5. ✅ `scripts/requirements.txt`

### 로컬 전용으로 유지:
1. ❌ `docker/docker-compose.yaml` (로컬 버전으로 복사 후 원본 복원)
2. ❌ 모든 `.txt`, `.log`, `.sql` 파일 (루트)
3. ❌ `scripts/.env` 및 임시 스크립트들

### .gitignore 업데이트:
```
# Local development
docker/docker-compose.local.yaml
scripts/.env

# Temporary files
*.log
sdk_trace.log
verify.log

# Local SQL scripts
/*.sql

# Local data dumps
/*.txt
/*.json
```
