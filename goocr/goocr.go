package goocr

type Config struct {
	credentialsFilePath string
	tokenFilePath       string
}

func NewConfig(credentialsFilePath, tokenFilePath string) *Config {
	return &Config{
		credentialsFilePath: credentialsFilePath,
		tokenFilePath:       tokenFilePath,
	}
}

type Goocr struct {
	config *Config
}

func NewGoocr(config *Config) *Goocr {
	return &Goocr{
		config: config,
	}
}
