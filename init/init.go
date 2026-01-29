package init

import "github.com/matheushermes/wedding_planner_service/configs"

func init() {
	configs.LoadEnv()
}