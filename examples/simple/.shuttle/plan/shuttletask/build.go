package shuttletask

import "context"

func Build(ctx context.Context) error {
	println("parent build")

	return nil
}
