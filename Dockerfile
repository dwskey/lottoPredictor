# --- 1단계: 빌드 ---
    FROM golang:1.24.2 AS builder

    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod tidy
    
    COPY . .
    RUN go build -o lottoApp main.go
    
    # --- 2단계: 실행 환경 (슬림화 가능) ---
    FROM debian:bullseye-slim
    WORKDIR /app
    
    # 실행에 필요한 파일만 복사
    COPY --from=builder /app/lottoApp .
    COPY config/ ./config/
    COPY database/ ./database/
    
    CMD ["./lottoApp"]
    