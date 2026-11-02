module trade-tools

go 1.26

require (
	github.com/SolaceDev/solace-agent-mesh-go/pkg/samtoolsdk v0.0.0
	github.com/jackc/pgx/v5 v5.11.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/text v0.29.0 // indirect
)

// TODO: drop the replace once samtoolsdk is consumable as a stable module.
replace github.com/SolaceDev/solace-agent-mesh-go/pkg/samtoolsdk => ./_sdk/samtoolsdk
