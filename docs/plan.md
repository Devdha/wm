# WM 2.0 기획서

## 목표
- `git worktree add`와 정렬된 UX로 `wm init/add/list/remove` 전 사이클 관리 제공
- 일관된 명령 체계, 안전한 삭제, 명확한 리스트 뷰, 선택형 자동화(sync, bg)

## 문제 정의
- 워크트리 생성/정리/목록/동기화 흐름이 분산되어 반복적이고 실수(브랜치 정리 누락) 가능

## 범위
### 포함 (2.0)
- `wm init`: 프로젝트 스캔 후 `.wm.yaml` 생성(TUI)
- `wm add`: 워크트리 생성 + 파일 sync + bg 작업
- `wm list`: 관리 중 워크트리 목록을 표로 출력
- `wm remove`: 워크트리 삭제 + 옵션(`-b`, `-f`)
- 기본 경로 규칙: `<repo_root>/..` 아래 `wm_<repo_root_name>/` 생성 후 그 안에 브랜치별 워크트리 생성
- `--path` 옵션으로 워크트리 베이스 경로 지정
- 크로스 플랫폼 지원(Windows/macOS/Linux)
- 모노레포 의존성 설치 감지 및 적절한 커맨드 추천/선택

### 제외 (2.0)
- 원격 브랜치 자동 관리
- 외부 이슈 트래커 연동
- 복잡한 템플릿 엔진

## 사용자 플로우
1. `wm init` -> `.wm.yaml` 생성
2. `wm add <branch>` -> 기본 베이스 경로에 worktree 생성 + sync + bg
3. `wm list` -> 관리중 worktree 표 출력
4. `wm remove <path> [-b] [-f]` -> worktree 제거 + (선택) 브랜치 삭제

## 리스크 및 완화
- 브랜치 삭제 오작동: 삭제 전 `git worktree list --porcelain`로 매칭 필수
- `remove -b` 안전장치: 다른 워크트리에서 사용 중인 브랜치는 삭제 금지
- 강제 삭제 오남용: 기본 확인 프롬프트, `-f`에서만 생략
- 경로/브랜치 매핑 실패: 메타 캐시 또는 즉시 조회 후 제거
- 플랫폼 차이: symlink/백그라운드 프로세스/알림은 OS별 분기
- 스캔 성능: 기본 ignore 목록 + 사용자 ignore 옵션 제공

## 산출물
- CLI 스켈레톤 + 명령별 핵심 로직
- `.wm.yaml` 스키마 및 샘플
- 사용 문서(README, 예시)

## 성공 기준
- `add/list/remove`가 표준 git worktree 대비 동일/더 나은 UX 제공
- `remove -b`가 안전하게 브랜치 삭제 처리
- 최소 5개 E2E 시나리오 통과

---

# 실행 체크리스트

## 1. 설계/스키마
- [ ] `wm init` TUI 플로우 정의
- [ ] `.wm.yaml` 스키마 확정
- [ ] `.wm.yaml` 샘플 작성

## 2. 핵심 기능
- [ ] 기본 워크트리 베이스 경로 규칙 확정 (`../wm_<repo_root_name>/`)
- [ ] `wm add --path` 옵션 설계 및 우선순위 정의
- [ ] `wm add` worktree 생성 성공/실패 처리
- [ ] 파일 sync 정책 확정 (문자열/객체 혼용, glob, overwrite, symlink)
- [ ] 스캔 ignore 기본 목록 및 사용자 ignore 옵션 확정
- [ ] bg 작업 실행 옵션 설계 (on/off, async)
- [ ] 모노레포 의존성 감지 및 커맨드 선택 로직 확정

## 3. 리스트/삭제
- [ ] `wm list` 파싱 안정화 (`git worktree list --porcelain`)
- [ ] `wm list` 표 출력 포맷 확정
- [ ] `wm remove` 안전 확인(기본) + `-f` 처리
- [ ] `wm remove -b` 브랜치 조회/삭제 순서 구현
- [ ] `wm remove -b` 안전장치(사용 중 브랜치 삭제 금지) 구현
- [ ] 경로-브랜치 매칭 실패 시 에러 메시지/가이드

## 4. 문서/테스트
- [ ] E2E 시나리오 테스트 5개 작성
- [ ] README 사용 예시 및 옵션 설명 추가
