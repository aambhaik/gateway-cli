package app

import (
	"encoding/json"
	"strings"

	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	ftrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	fflow "github.com/TIBCOSoftware/flogo-lib/flow/flowdef"
	"github.com/aambhaik/gateway-cli/env"
	"github.com/aambhaik/gateway-cli/types"
	"github.com/aambhaik/gateway-cli/util"
	"errors"
	factions "github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/app"
	fapi "github.com/TIBCOSoftware/flogo-cli/app"
)

var flowActivitiesMap map[string] (map[string] int)

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

	triggerNamedMap := make(map[string] types.Trigger)
	for _, trigger := range descriptor.Gateway.Triggers {
		triggerNamedMap[trigger.Name] = trigger
	}

	handlerNamedMap := make(map[string] types.EventHandler)
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

			var flow = types.Flow{
				Type: 1,
				RootTask: flogoAction,
			}

			var flowData = types.FlowData{
				Flow: flow,
			}

			data, err := json.Marshal(flowData)

			actionObject := factions.Config{
				Id: handlerName,
				Ref: "github.com/TIBCOSoftware/flogo-contrib/action/flow",
				Data: data,
			}

			flogoAppActions = append(flogoAppActions, &actionObject)
		}
	}

	flogoApp := app.Config{
		Name: descriptor.Gateway.Name,
		Type: "flogo:app",
		Version: descriptor.Gateway.Version,
		Description:descriptor.Gateway.Description,
		Triggers: flogoAppTriggers,
		Actions: flogoAppActions,
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

func CreateFlogoFlowAction(handler types.EventHandler) (*fflow.TaskRep, error) {
	var fflowAction fflow.TaskRep

	fflowAction.Name = handler.Name

	var flowTasks []*fflow.TaskRep
	var flowLinks []*fflow.LinkRep

	for _, action := range handler.Actions {
		flowTask, err := CreateFlogoActivity(action)
		if err != nil {
			return nil, err
		}

		flowTasks = append(flowTasks, flowTask)
	}

	for _, link := range handler.Links {
		flowLink, err := CreateFlogoActionLink(handler.Name, link)
		if err != nil {
			return nil, err
		}
		flowLinks = append(flowLinks, flowLink)
	}

	rootTask := fflow.TaskRep{
		ID: 1,
		TypeID: 1,
		Tasks: flowTasks,
		Links: flowLinks,
	}

	return &rootTask, nil
}

func CreateFlogoActivity(action types.Action) (*fflow.TaskRep, error) {
	var flowTask fflow.TaskRep
	flowTask.Name = action.Name
	flowTask.ID = action.ID
	flowTask.ActivityRef = action.Type
	var actionInputs interface{}
	if err := json.Unmarshal([]byte(action.Inputs), &actionInputs); err != nil {
		panic(err)
	}

	actionInputsMap := actionInputs.(map[string]interface{})

	var taskAttributes []*data.Attribute

	for inputName := range actionInputsMap {
		inputValue := data.GetMapValue(actionInputsMap, inputName)
		inputType, err := data.GetType(inputValue)
		if err != nil {
			return nil, err
		}

		taskAttributes = append(taskAttributes,
			&data.Attribute{
				Name:  inputName,
				Type:  inputType,
				Value: inputValue,
			})
	}
	flowTask.Attributes = taskAttributes

	return &flowTask, nil
}

func CreateFlogoActionLink(flowName string, link types.Link) (*fflow.LinkRep, error) {
	var flowLink fflow.LinkRep

	flowActionIdMap := flowActivitiesMap[flowName];
	if flowActionIdMap == nil {
		err := errors.New("invalid flow name input")
		return nil, err
	}

	fromId := flowActionIdMap[link.From]
	if &fromId == nil {
		err := errors.New("invalid action name in the link")
		return nil, err
	}
	flowLink.FromID = fromId

	toId := flowActionIdMap[link.To]
	if &toId == nil {
		err := errors.New("invalid action name in the link")
		return nil, err
	}
	flowLink.ToID = toId

	flowLink.Value = link.If
	flowLink.Type = 1

	return &flowLink, nil
}

