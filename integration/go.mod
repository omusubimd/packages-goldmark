module github.com/omusubimd/packages-goldmark/integration

go 1.25

require (
	github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent v0.0.0
	github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby v0.0.0
	github.com/yuin/goldmark/v2 v2.1.5
)

replace github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent => ../packages/goldmark-omusubi-indent

replace github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby => ../packages/goldmark-omusubi-ruby
