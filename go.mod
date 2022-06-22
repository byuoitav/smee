module github.com/byuoitav/smee

go 1.15

replace github.com/byuoitav/av-cli => ../av-cli

require (
	github.com/byuoitav/auth v0.3.3
	github.com/byuoitav/av-cli v1.1.0
	github.com/byuoitav/central-event-system v0.0.0-20201020053146-aee08228b14a
	github.com/byuoitav/common v0.0.0-20191210190714-e9b411b3cc0d
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.8.0
	github.com/golang-jwt/jwt/v4 v4.4.1
	github.com/gwatts/gin-adapter v0.0.0-20170508204228-c44433c485ad
	github.com/jackc/pgx/v4 v4.11.0
	github.com/matryer/is v1.4.0
	github.com/segmentio/ksuid v1.0.3
	github.com/spf13/pflag v1.0.5
	github.com/valyala/fasttemplate v1.1.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.27.1
)
