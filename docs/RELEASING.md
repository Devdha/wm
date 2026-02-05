# Releasing

wm 릴리스 프로세스 가이드.

## 배포 구조

```
gitwm (메인 패키지)
├── bin/wm (플랫폼 감지 → 바이너리 실행)
└── optionalDependencies:
    ├── gitwm-darwin-arm64
    ├── gitwm-darwin-x64
    ├── gitwm-linux-arm64
    ├── gitwm-linux-x64
    └── gitwm-win32-x64
```

- GoReleaser로 5개 플랫폼용 바이너리 빌드
- 각 플랫폼별 npm 패키지 + 메인 패키지 (총 6개)
- `npm install -g gitwm` 시 현재 플랫폼에 맞는 바이너리만 설치됨

## 릴리스 방법

### 1. 태그 생성 및 푸시

```bash
git tag v0.2.0
git push origin v0.2.0
```

GitHub Actions가 자동으로:
1. GoReleaser로 바이너리 빌드
2. GitHub Release 생성
3. npm publish 시도 (2FA 계정은 실패함)

### 2. npm 수동 배포 (2FA 계정)

npm 2FA가 활성화된 경우 로컬에서 수동 배포 필요:

```bash
# npm 로그인 (처음 한 번만)
npm login

# 로컬 배포 스크립트 실행
./scripts/publish-npm-local.sh 0.2.0
```

OTP를 6번 입력해야 함 (플랫폼 패키지 5개 + 메인 패키지 1개).

### 3. CI에서 자동 배포 (2FA 없는 계정)

2FA가 없거나 Automation 토큰이 있는 경우 GitHub Actions에서 자동 배포됨.

Repository secrets에 `NPM_TOKEN` 설정 필요:

```bash
gh secret set NPM_TOKEN
```

## 버전 관리

- Semantic Versioning 사용: `MAJOR.MINOR.PATCH`
- 태그 형식: `v0.1.0`, `v1.0.0`

## 설치 확인

```bash
# npm
npm install -g gitwm
wm --version

# Go
go install github.com/Devdha/wm@latest
wm --version
```

## 트러블슈팅

### npm 403 Forbidden
- Granular Access Token 권한 확인
- "Allow this token to publish new packages" 활성화 필요

### npm EOTP
- 2FA 계정은 CI에서 자동 배포 불가
- `./scripts/publish-npm-local.sh` 사용

### 패키지 이미 존재
- 동일 버전 재배포 불가
- 버전 bump 후 다시 시도
