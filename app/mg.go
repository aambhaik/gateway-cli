package app

import (
	"encoding/json"
	"strings"

	"fmt"
	fapi "github.com/TIBCOSoftware/flogo-cli/app"
	"github.com/TIBCOSoftware/flogo-lib/app"
	factions "github.com/TIBCOSoftware/flogo-lib/core/action"
	ftrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/aambhaik/gateway-cli/env"
	"github.com/aambhaik/gateway-cli/types"
	"github.com/aambhaik/gateway-cli/util"
)

var flowActivitiesMap map[string](map[string]int)

// CreateGateway creates a gateway application from the specified json gateway descriptor
func CreateGateway(env env.Project, gatewayJson string, appDir string, appName string, vendorDir string) error {

	descriptor, err := ParseGatewayDescriptor(gatewayJson)
	if err != nil {
		return err
	}

	if appName != "" {
		appName := appName + "_new"
		// override the application name

		altJson := strings.Replace(gatewayJson, `"`+descriptor.Gateway.Name+`"`, `"`+appName+`"`, 1)
		altDescriptor, err := ParseGatewayDescriptor(altJson)

		//see if we can get away with simple replace so we don't reorder the existing json
		if err == nil && altDescriptor.Gateway.Name == appName {
			gatewayJson = altJson
		} else {
			//simple replace didn't work so we have to unmarshal & re-marshal the supplied json
			var appObj map[string]interface{}
			err := json.Unmarshal([]byte(gatewayJson), &appObj)
			if err != nil {
				return err
			}

			appObj["name"] = appName

			updApp, err := json.MarshalIndent(appObj, "", "  ")
			if err != nil {
				return err
			}
			gatewayJson = string(updApp)
		}

		descriptor.Gateway.Name = appName
	}

	env.Init(appDir)
	err = env.Create(false, vendorDir)
	if err != nil {
		return err
	}

	err = fgutil.CreateFileFromString(fgutil.Path(appDir, "gateway.json"), gatewayJson)
	if err != nil {
		return err
	}

	flogoAppTriggers := []*ftrigger.Config{}
	flogoAppActions := []*factions.Config{}

	triggerNamedMap := make(map[string]types.Trigger)
	for _, trigger := range descriptor.Gateway.Triggers {
		triggerNamedMap[trigger.Name] = trigger
	}

	handlerNamedMap := make(map[string]types.EventHandler)
	for _, evtHandler := range descriptor.Gateway.EventHandlers {
		handlerNamedMap[evtHandler.Name] = evtHandler
	}

	//translate the gateway model to the flogo model
	for _, link := range descriptor.Gateway.EventLinks {
		triggerName := link.Trigger

		successPaths := link.SuccessPaths
		for _, path := range successPaths {
			handlerName := path.Handler

			flogoTrigger, err := CreateFlogoTrigger(triggerNamedMap[triggerName], handlerNamedMap[handlerName])
			if err != nil {
				return err
			}

			flogoAppTriggers = append(flogoAppTriggers, flogoTrigger)

			flogoAction, err := CreateFlogoFlowAction(handlerNamedMap[handlerName])
			if err != nil {
				return err
			}

			flogoAppActions = append(flogoAppActions, flogoAction)
		}
	}

	flogoApp := app.Config{
		Name:        descriptor.Gateway.Name,
		Type:        "flogo:app",
		Version:     descriptor.Gateway.Version,
		Description: descriptor.Gateway.Description,
		Triggers:    flogoAppTriggers,
		Actions:     flogoAppActions,
	}

	//create flogo PP JSON
	bytes, err := json.MarshalIndent(flogoApp, "", "\t")
	if err != nil {
		return nil
	}

	flogoJson := string(bytes)
	//fmt.Printf("flogoJson: %s \n", flogoJson)
	err = fgutil.CreateFileFromString(fgutil.Path(appDir, "flogo.json"), flogoJson)
	if err != nil {
		return err
	}
	fmt.Printf("Generated flogo JSON in %s \n", fgutil.Path(appDir, "flogo.json"))

	fapi.CreateApp(env, flogoJson, appDir, appName, vendorDir)
	fmt.Println("Generated flogo Artifacts.")

	return nil
}

// ParseGatewayDescriptor parse the application descriptor
func ParseGatewayDescriptor(appJson string) (*types.Microgateway, error) {
	descriptor := &types.Microgateway{}

	err := json.Unmarshal([]byte(appJson), descriptor)

	if err != nil {
		return nil, err
	}

	return descriptor, nil
}

func CreateFlogoTrigger(trigger types.Trigger, handler types.EventHandler) (*ftrigger.Config, error) {
	var flogoTrigger ftrigger.Config
	flogoTrigger.Name = trigger.Name
	flogoTrigger.Id = trigger.Name
	flogoTrigger.Ref = trigger.Type
	var ftSettings interface{}
	if err := json.Unmarshal([]byte(trigger.Settings), &ftSettings); err != nil {
		return nil, err
	}
	flogoTrigger.Settings = ftSettings.(map[string]interface{})
	flogoHandler := ftrigger.HandlerConfig{
		ActionId: handler.Name,
		Settings: ftSettings.(map[string]interface{}),
	}

	handlers := []*ftrigger.HandlerConfig{}
	handlers = append(handlers, &flogoHandler)

	flogoHandler.Settings["useReplyHandler"] = "true"
	flogoHandler.Settings["autoIdReply"] = "true"
	flogoTrigger.Handlers = handlers

	return &flogoTrigger, nil
}

func CreateFlogoFlowAction(handler types.EventHandler) (*factions.Config, error) {
	flogoAction := types.FlogoAction{}
	err := json.Unmarshal([]byte(handler.Definition), &flogoAction)
	if err != nil {
		return nil, err
	}
	action := factions.Config{
		Id:   handler.Name,
		Data: flogoAction.Data,
		Ref:  flogoAction.Ref,
	}
	return &action, nil
}
