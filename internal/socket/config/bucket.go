package config

type Bucket struct {
	BucketCount      int64 `mapstructure:"BucketCount"`
	JobChanSize      int64 `mapstructure:"JobChanSize"`
	DispatchChanSize int64 `mapstructure:"DispatchChanSize"`
	BucketFanOutCount int64 `mapstructure:"BucketFanOutCount"`
}