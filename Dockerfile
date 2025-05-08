# 베이스 이미지: 공식 Go 이미지 (최신 안정 버전)
FROM golang:1.24.2

# 작업 디렉토리 생성
WORKDIR /app

# 모듈 설정 파일 복사
COPY go.mod go.sum ./

# 의존성 설치 (캐시 최적화용)
RUN go mod tidy

# 전체 소스코드 복사
COPY . .

# 빌드 (main.go를 기준으로)
RUN go build -o lottoApp main.go

# 컨테이너 실행 시 동작
CMD ["./lottoApp"]

