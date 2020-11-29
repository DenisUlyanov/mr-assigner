package cmd

import (
	"context"
	"github.com/DenisUlyanov/mr-assigner/internal/handlers"
	"github.com/DenisUlyanov/mr-assigner/internal/services/gitlab"
	"github.com/spf13/cobra"
	"time"
)

var serviceStartAssignerCmd = &cobra.Command{
	Use:   "assigner",
	Short: "starts the assigner server",
	Run: func(cmd *cobra.Command, args []string) {
		errc := make(chan error)
		ctx, cancelFunc := context.WithCancel(context.Background())

		go handlers.InterruptHandler(errc, cancelFunc)

		go func(ctx context.Context, d time.Duration) {
			for range time.Tick(d) {
				func(ctx context.Context) {
					gitlab.CreateGitlabService("", "").AssignMergeRequest(ctx)
				}(ctx)
			}
		}(ctx, time.Second)

		<-ctx.Done()
	},
}

func init() {
	serviceCmd.AddCommand(serviceStartAssignerCmd)
}
