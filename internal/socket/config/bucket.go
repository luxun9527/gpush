package config

type Bucket struct {
	BucketCount      int64 `mapstructure:"BucketCount"`
	DispatchChanSize int64 `mapstructure:"DispatchChanSize"`
}
