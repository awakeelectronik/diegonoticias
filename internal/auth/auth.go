package auth

import "time"

const SessionCookieName = "dn_session"

func SessionTTL() time.Duration {
	return 7 * 24 * time.Hour
}

