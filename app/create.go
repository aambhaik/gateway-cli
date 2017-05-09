package app

import (
	"flag"
	"fmt"
	"os"

	"encoding/json"
	"github.com/aambhaik/gateway-cli/cli"
	"github.com/aambhaik/gateway-cli/types"
	"github.com/aambhaik/gateway-cli/util"
)

var optCreate = &cli.OptionInfo{
	Name:      "create",
	UsageLine: "create AppName",
	Short:     "create a gateway project",
	Long: `Creates a gateway project.

Options:
    -f       specify the gateway.json to create project from
    -vendor  specify existing vendor directory to copy

 `,
}

func init() {
	CommandRegistry.RegisterCommand(&cmdCreate{option: optCreate})
}

type cmdCreate struct {
	option    *cli.OptionInfo
	fileName  string
	vendorDir string
}

// HasOptionInfo implementation of cli.HasOptionInfo.OptionInfo
func (c *cmdCreate) OptionInfo() *cli.OptionInfo {
	return c.option
}

// AddFlags implementation of cli.Command.AddFlags
func (c *cmdCreate) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&(c.fileName), "f", "", "gateway app file")
	fs.StringVar(&(c.vendorDir), "vendor", "", "vendor dir")
}

// Exec implementation of cli.Command.Exec
func (c *cmdCreate) Exec(args []string) error {

	var gatewayJson string
	var gatewayName string
	var err error

	if c.fileName != "" {

		if fgutil.IsRemote(c.fileName) {

			gatewayJson, err = fgutil.LoadRemoteFile(c.fileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Error loading app file '%s' - %s\n\n", c.fileName, err.Error())
				os.Exit(2)
			}
		} else {
			gatewayJson, err = fgutil.LoadLocalFile(c.fileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Error loading app file '%s' - %s\n\n", c.fileName, err.Error())
				os.Exit(2)
			}

			if len(args) != 0 {
				gatewayName = args[0]
			}
		}
	} else {
		if len(args) == 0 {
			fmt.Fprint(os.Stderr, "Error: Gateway name not specified\n\n")
			cmdUsage(c)
		}

		if len(args) != 1 {
			fmt.Fprint(os.Stderr, "Error: Too many arguments given\n\n")
			cmdUsage(c)
		}

		gatewayName = args[0]
		microGateway, err := createMicroGatewayModel()
		//microGateway, err := createMicroGatewayModelKafka()
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(microGateway, "", "\t")
		if err != nil {
			return err
		}
		gatewayJson = string(bytes)
	}

	currentDir, err := os.Getwd()

	if err != nil {
		return err
	}

	appDir := fgutil.Path(currentDir, gatewayName)

	return CreateGateway(SetupNewProjectEnv(), gatewayJson, appDir, gatewayName, c.vendorDir)
}

func createMicroGatewayModel() (types.Microgateway, error) {

	microGateway := types.Microgateway{
		Gateway: types.Gateway{
			Name:           "Test",
			Version:        "1.0.0",
			Description:    "This is the first microgateway app",
			Configurations: []types.Config{},
			Triggers: []types.Trigger{
				{
					Name:        "rest_trigger",
					Description: "The trigger on 'users' endpoint",
					Type:        "github.com/TIBCOSoftware/flogo-contrib/trigger/rest",
					Settings: json.RawMessage(`{
					  "port": "9096",
					  "method": "GET",
					  "path": "/v1/mg/hello"
					}`),
				},
			},
			EventHandlers: []types.EventHandler{
				{
					Name:        "get_user_success_handler",
					Description: "Handle the user access",
					Params: json.RawMessage(`{
                    				"internalUrl": "hello"
					}`),
					Definition: json.RawMessage(`{
					                    "data": {
								"flow": {
								    "type": 1,
								    "attributes": [],
								    "rootTask": {
									"id": 1,
									"type": 1,
									"tasks": [
									    {
										"id": 2,
										"name": "Invoke REST Service",
										"description": "Simple REST Activity",
										"type": 1,
										"activityType": "tibco-rest",
										"activityRef": "github.com/TIBCOSoftware/flogo-contrib/activity/rest",
										"attributes": [
										    {
											"name": "method",
											"value": "GET",
											"required": true,
											"type": "string"
										    },
										    {
											"name": "uri",
											"value": "{internalUrl}",
											"required": true,
											"type": "string"
										    },
										    {
											"name": "pathParams",
											"value": null,
											"required": false,
											"type": "params"
										    },
										    {
											"name": "queryParams",
											"value": null,
											"required": false,
											"type": "params"
										    },
										    {
											"name": "content",
											"value": null,
											"required": false,
											"type": "any"
										    }
										]
									    },
									    {
										"id": 3,
										"name": "Log Message",
										"description": "Simple Log Activity",
										"type": 1,
										"activityType": "tibco-log",
										"activityRef": "github.com/TIBCOSoftware/flogo-contrib/activity/log",
										"attributes": [
										    {
											"name": "message",
											"value": "",
											"required": false,
											"type": "string"
										    },
										    {
											"name": "flowInfo",
											"value": "false",
											"required": false,
											"type": "boolean"
										    },
										    {
											"name": "addToFlow",
											"value": "false",
											"required": false,
											"type": "boolean"
										    }
										]
									    },
									    {
										"id": 4,
										"name": "Reply To Trigger",
										"description": "Simple Reply Activity",
										"type": 1,
										"activityType": "tibco-reply",
										"activityRef": "github.com/TIBCOSoftware/flogo-contrib/activity/reply",
										"attributes": [
										    {
											"name": "code",
											"value": null,
											"required": true,
											"type": "integer"
										    },
										    {
											"name": "data",
											"value": null,
											"required": false,
											"type": "any"
										    }
										],
										"inputMappings": [
										    {
											"type": 1,
											"value": "{A2.result}.status",
											"mapTo": "code"
										    },
										    {
											"type": 1,
											"value": "{A2.result}",
											"mapTo": "data"
										    }
										]
									    }
									],
									"links": [
									    {
										"id": 1,
										"from": 2,
										"to": 3,
										"type": 0
									    },
									    {
										"id": 2,
										"from": 3,
										"to": 4,
										"type": 0
									    }
									],
									"attributes": []
								    }
								}
							    },
							    "id": "get_user_success_handler",
							    "ref": "github.com/TIBCOSoftware/flogo-contrib/action/flow"
					}`),
				},
			},
			EventLinks: []types.EventLink{
				{
					Trigger: "rest_trigger",
					SuccessPaths: []types.Path{
						{
							Handler: "get_user_success_handler",
						},
					},
					ErrorPaths: []types.Path{},
				},
			},
		},
	}

	return microGateway, nil
}

//func createMicroGatewayModelKafka() (types.Microgateway, error) {
//
//	microGateway := types.Microgateway{
//		Gateway: types.Gateway{
//			Name:        "Test",
//			Version:     "1.0.0",
//			Description: "This is the first microgateway app",
//			Configurations: []types.Config{
//				{
//					Name:        "kafkaConfig",
//					Type:        "github.com/TIBCOSoftware/flogo-contrib/config/kafkaConfig",
//					Description: "Configuration for kafka cluster",
//					Settings: json.RawMessage(`{
//						"brokers": [
//						 "localhost:9092",
//						 "localhost:9093"
//						],
//						"userName": "admin",
//						"password": "admin"
//					}`),
//				},
//			},
//			Triggers: []types.Trigger{
//				{
//					Name:        "OrdersTrigger",
//					Description: "The trigger on 'orders' topic",
//					Type:        "github.com/TIBCOSoftware/flogo-contrib/trigger/kafkaConsumer",
//					Settings: json.RawMessage(`{
//					  "topic": "orders",
//					  "config": "${configurations.kafkaConfig}"
//					}`),
//				},
//			},
//			EventHandlers: []types.EventHandler{
//				{
//					Name:        "OrderSuccessHandler",
//					Description: "Handle the order processing",
//					Params: json.RawMessage(`{
//					  "time-span": "minute",
//					  "message": "${trigger.content}",
//					  "limit": 1001
//					}`),
//					Actions: []types.Action{
//						{
//							ID:          1,
//							Name:        "log",
//							Description: "Log the inbound request",
//							Type:        "github.com/TIBCOSoftware/flogo-contrib/activity/log",
//							Inputs: json.RawMessage(`{
//								"message": {
//									"value": "${inputs.message}",
//									"type": "string"
//							    	}
//							}`),
//						},
//						{
//							ID:          2,
//							Name:        "limit",
//							Description: "Limit the traffic to endpoint",
//							Type:        "github.com/TIBCOSoftware/flogo-contrib/activity/RateLimiter",
//							Inputs: json.RawMessage(`{
//								"operation": "POST http: //localhost:9090/users",
//								"limit": "${inputs.limit}",
//								"time-span": "${inputs.time-span}"
//							}`),
//							Outputs: json.RawMessage(`{
//								"operation": {
//									"type": "string"
//								},
//								"throttled": {
//									"type": "boolean"
//								},
//								"error": {
//									"type": "error"
//								}
//							}`),
//						},
//						{
//							ID:          3,
//							Name:        "invoke",
//							Description: "Invoke the endpoint",
//							Type:        "github.com/TIBCOSoftware/flogo-contrib/activity/RESTInvoke",
//							Inputs: json.RawMessage(`{
//								"operation": "${action.limit.operation}",
//								"content": "${inputs.message}"
//							}`),
//							Outputs: json.RawMessage(`{
//								"status": {
//									"type": "string"
//								},
//								"message": {
//									"type": "json"
//								},
//								"error": {
//									"type": "error"
//								}
//							}`),
//						},
//						{
//							ID:          4,
//							Name:        "error",
//							Description: "Error handling",
//							Type:        "github.com/TIBCOSoftware/flogo-contrib/activity/error",
//							Inputs: json.RawMessage(`{
//								"message": "No more than ${inputs.limit} requests allowed per ${inputs.time-span}"
//							}`),
//						},
//					},
//					Links: []types.Link{
//						{
//							From: "limit",
//							To:   "error",
//							If:   "${limit.throttled} == true",
//						},
//						{
//							From: "limit",
//							To:   "invoke",
//							If:   "${limit.throttled} == false",
//						},
//					},
//				},
//				{
//					Name:        "OrderErrorHandler",
//					Description: "Handle the order error processing",
//					Reference:   "github.com/TIBCOSoftware/mashling/flows/ConsoleErrorHandler/flow.json",
//					Params: json.RawMessage(`{
//					  "message": "${trigger.content}"
//					}`),
//				},
//			},
//			EventLinks: []types.EventLink{
//				{
//					Trigger: "OrdersTrigger",
//					SuccessPaths: []types.Path{
//						{
//							If:      "${trigger.content.Country == \"USA\"}",
//							Handler: "OrderSuccessHandler",
//						},
//					},
//					ErrorPaths: []types.Path{
//						{
//							If:      "${trigger.content.Country == undefined}",
//							Handler: "OrderErrorHandler",
//						},
//					},
//				},
//			},
//		},
//	}
//
//	return microGateway, nil
//}
