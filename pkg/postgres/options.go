package postgres

import "time"

type Option func(*Postgres)

func MaxPoolSize(s int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = s
	}
}

func ConnAttemps(c int) Option {
	return func(p *Postgres) {
		p.connAttemps = c
	}
}

func ConnTimeout(timeOut time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = timeOut
	}
}
