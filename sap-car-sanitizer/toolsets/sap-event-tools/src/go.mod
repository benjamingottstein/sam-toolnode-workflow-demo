module sap-event-tools

go 1.26

require (
	github.com/SolaceDev/solace-agent-mesh-go/pkg/samtoolsdk v0.0.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
)

require golang.org/x/text v0.14.0 // indirect

// TODO: drop the replace once samtoolsdk is consumable as a stable module.
replace github.com/SolaceDev/solace-agent-mesh-go/pkg/samtoolsdk => ./_sdk/samtoolsdk
