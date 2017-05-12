package app

import (
	"encoding/json"
	"strings"

	"fmt"
	fapi "github.com/TIBCOSoftware/flogo-cli/app"
	"github.com/TIBCOSoftware/flogo-lib/app"
	factions "github.com/TIBCOSoftware/flogo-lib/core/action"
	ftrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/flow/flowdef"
	"github.com/aambhaik/gateway-cli/env"
	"github.com/aambhaik/gateway-cli/types"
	"github.com/aambhaik/gateway-cli/util"
	"github.com/pkg/errors"
	"io/ioutil"
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

	//env.Init(appDir)
	//err = env.Create(false, vendorDir)
	//if err != nil {
	//	return err
	//}

	//gatewayConfigurations := []*types.Config{}
	flogoAppTriggers := []*ftrigger.Config{}
	flogoAppActions := []*factions.Config{}

	configNamedMap := make(map[string]types.Config)
	for _, config := range descriptor.Gateway.Configurations {
		configNamedMap[config.Name] = config
	}


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

	fapi.CreateApp(SetupNewProjectEnv(), flogoJson, appDir, appName, vendorDir)
	//err = fgutil.CreateFileFromString(fgutil.Path(appDir, "gateway.json"), gatewayJson)
	//if err != nil {
	//	return err
	//}
	//
	//err = fgutil.CreateFileFromString(fgutil.Path(appDir, "flogo.json"), flogoJson)
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("Generated flogo JSON in %s \n", fgutil.Path(appDir, "flogo.json"))

	fmt.Println("Generated gateway Artifacts.")
	fmt.Println("Building gateway Artifacts.")

	options := &fapi.BuildOptions{SkipPrepare: false, PrepareOptions: &fapi.PrepareOptions{OptimizeImports: false, EmbedConfig: false}}
	fapi.BuildApp(SetupExistingProjectEnv(appDir), options)

	err = fgutil.CreateFileFromString(fgutil.Path(appDir, "gateway.json"), gatewayJson)
	if err != nil {
		return err
	}

	fmt.Println("Gateway successfully built!")

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

func CreateGatewayConfiguration(trigger types.Trigger, handler types.EventHandler) (*ftrigger.Config, error) {
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
	reference := &handler.Reference
	gatewayAction := factions.Config{}

	if reference != nil {
		//reference is provided, get the referenced resource inline. the provided path should be the git path e.g. github.com/aambhaik/resources/app.json
		referenceString := *reference

		index := strings.LastIndex(referenceString, "/")

		if index < 0 {
			return nil, errors.New("Invalid URL reference. Pls provide the github path to mashling pattern flow json")
		}
		gitHubPath := referenceString[0:index]

		resourceFile := referenceString[index+1 : len(referenceString)]

		gbProject := env.NewGbProjectEnv()

		err := gbProject.InstallDependency(gitHubPath, "")
		if err != nil {
			return nil, err
		}
		resourceDir := gbProject.GetVendorSrcDir()
		resourcePath := resourceDir + "/" + gitHubPath + "/" + resourceFile
		data, err := ioutil.ReadFile(resourcePath)
		if err != nil {
			return nil, err
		}

		var flogoFlowDef *app.Config
		err = json.Unmarshal(data, &flogoFlowDef)
		if err != nil {
			return nil, err
		}

		actions := flogoFlowDef.Actions
		if len(actions) != 1 {
			return nil, errors.New("Please make sure that the pattern flow has only one action")
		}

		action := actions[0]
		action.Id = handler.Name
		gatewayAction = factions.Config{
			Id:   handler.Name,
			Data: action.Data,
			Ref:  action.Ref,
		}

	} else if handler.Definition != nil {
		//definition is provided inline
		err := json.Unmarshal([]byte(handler.Definition), &flogoAction)
		if err != nil {
			return nil, err
		}
		gatewayAction = factions.Config{
			Id:   handler.Name,
			Data: flogoAction.Data,
			Ref:  flogoAction.Ref,
		}
	}

	return &gatewayAction, nil

	//NewAction, err := ReplaceActionParams(&gatewayAction, handler)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return NewAction, nil
}

func ReplaceActionParams(action *factions.Config, handler types.EventHandler) (*factions.Config, error) {
	var floDef = flowdef.DefinitionRep{}
	err := json.Unmarshal([]byte(action.Data), &floDef)
	if err != nil {
		return nil, err
	}
	var paramMap map[string]interface{}
	err = json.Unmarshal(handler.Params, &paramMap)
	if err != nil {
		return nil, err
	}
	for key, val := range paramMap {
		for _, task := range floDef.RootTask.Tasks {
			for _, attribute := range task.Attributes {
				if attribute.Name == key {
					attribute.Value = val
				}
			}
		}
	}
	modFlowDef, err := json.Marshal(&floDef)
	if err != nil {
		return nil, err
	}

	action.Data = modFlowDef

	return action, nil
}
