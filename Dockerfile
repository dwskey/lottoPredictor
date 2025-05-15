# Go 1.24.2 베이스 이미지
FROM golang:1.24.2

# 작업 디렉토리 설정
WORKDIR /lottopredictor

# 모듈 정보 복사 → 의존성 먼저 설치
COPY go.mod go.sum ./
RUN go mod tidy

# 전체 소스 복사
COPY . .

# 기본 실행 명령어 (개발환경에서는 sleep)
CMD ["sleep", "infinity"]
