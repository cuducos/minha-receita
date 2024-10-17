package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cuducos/go-cnpj"
)

const (
	retries           = 13
	timeoutPerAttempt = 1 * time.Second
)

var errTimeout = errors.New("getCompany timed out")

// this wrapper avoids having the getCompany idle for too long, wrapping it in
// timeout and restarting it after that
func getCompany(db database, n string) (string, error) {
	var c string
	err := retry.Do(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), timeoutPerAttempt)
			defer cancel()
			ch := make(chan error, 1)
			go func() {
				var err error
				c, err = db.GetCompany(cnpj.Unmask(n))
				ch <- err
			}()
			select {
			case <-ctx.Done():
				return errTimeout
			case err := <-ch:
				return err
			}
		},
		retry.Attempts(retries),
		retry.RetryIf(func(err error) bool {
			return err != nil && errors.Is(err, errTimeout)
		}),
	)
	if err != nil {
		return "", fmt.Errorf("error retrieving %s: %w", n, err)
	}
	return c, nil
}
