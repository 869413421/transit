package config

// ModelConfig 模型配置
type ModelConfig struct {
	Name                   string  `mapstructure:"name"`
	UpstreamName           string  `mapstructure:"upstream_name"`
	Type                   string  `mapstructure:"type"` // sync, async
	PricePer1KInputTokens  float64 `mapstructure:"price_per_1k_input_tokens"`
	PricePer1KOutputTokens float64 `mapstructure:"price_per_1k_output_tokens"`
	PricePerGeneration     float64 `mapstructure:"price_per_generation"`
}

// ModelsConfig 模型配置集合
type ModelsConfig struct {
	Text  []ModelConfig `mapstructure:"text"`
	Image []ModelConfig `mapstructure:"image"`
	Video []ModelConfig `mapstructure:"video"`
}

// AllModels 所有模型配置
type AllModels struct {
	Models ModelsConfig `mapstructure:"models"`
}

// GetModelByName 根据名称获取模型配置
func (m *ModelsConfig) GetModelByName(name string) *ModelConfig {
	// 在文本模型中查找
	for i := range m.Text {
		if m.Text[i].Name == name {
			return &m.Text[i]
		}
	}
	// 在图片模型中查找
	for i := range m.Image {
		if m.Image[i].Name == name {
			return &m.Image[i]
		}
	}
	// 在视频模型中查找
	for i := range m.Video {
		if m.Video[i].Name == name {
			return &m.Video[i]
		}
	}
	return nil
}
