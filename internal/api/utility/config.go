package utility

import (
	"flag"
	"github.com/lmittmann/tint"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port int
	ENV  string
	DB   struct {
		DSN             string
		MaxOpenConn     int
		MaxIdleConn     int
		MaxIdleConnTime string
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

func ParseFlags() *Config {
	var cfg Config
	flag.IntVar(&cfg.Port, "port", 8080, "API server Port")
	flag.StringVar(&cfg.ENV, "env", "dev", "Environment (dev|stag|prod)")
	// DB Flags
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.DB.MaxOpenConn, "db-max-open-conn", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConn, "db-max-idle-conn", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.DB.MaxIdleConnTime, "db-max-idle-time", "15m", "PostgreSQL max idle connection time")
	// SMTP Flags
	flag.StringVar(&cfg.SMTP.Host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP server host")
	flag.IntVar(&cfg.SMTP.Port, "smtp-port", 587, "SMTP server port")
	flag.StringVar(&cfg.SMTP.Username, "smtp-username", "d74896c632ff65", "SMTP username")
	flag.StringVar(&cfg.SMTP.Password, "smtp-password", "d79a98c2ff9076", "SMTP password")
	flag.StringVar(&cfg.SMTP.Sender, "smtp-sender", "Letschat <no-reply.letschat@muhammadusman.site>", "SMTP sender")
	flag.Parse()
	return &cfg
}

// ConfigureSlog so that it easy to locate the source file & line as the Goland IDE picks up the relative file path.
func ConfigureSlog(writeTo io.Writer) {
	wd, err := os.Getwd()
	var tintHandler slog.Handler
	if err != nil {
		slog.Error("Unable to find working dir, falling back to default slog Config")
		tintHandler = tint.NewHandler(writeTo, &tint.Options{AddSource: true})
	} else {
		unixPath := filepath.ToSlash(wd)
		tintHandler = tint.NewHandler(writeTo, &tint.Options{
			AddSource: true,
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.SourceKey {
					source, ok := attr.Value.Any().(*slog.Source)
					relativePath := "." + strings.TrimPrefix(source.File, unixPath)
					var sb strings.Builder
					sb.WriteString(relativePath)
					sb.WriteString(":")
					sb.WriteString(strconv.Itoa(source.Line))
					if !ok {
						panic("Unable to assert type on source attr while configuring tint handler")
					}
					return slog.Attr{
						Key:   attr.Key,
						Value: slog.StringValue(sb.String()),
					}
				}
				return attr
			},
		})
	}
	slog.SetDefault(slog.New(tintHandler))
}
