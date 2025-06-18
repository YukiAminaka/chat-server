FROM golang:1.24-alpine


# ホットリロードツール Air をインストール
RUN go install github.com/air-verse/air@latest


# コンテナ起動時に Air を実行する
CMD ["air", "-c", ".air.toml"]