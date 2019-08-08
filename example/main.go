package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"go.aporeto.io/elemental"
	"go.aporeto.io/manipulate"
	"go.uber.org/zap"

	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate/maniphttp"
	midgardclient "go.aporeto.io/midgard-lib/client"
	"go.aporeto.io/midgard-lib/tokenmanager"
)

const (
	// The appcred must be created ahead of time.
	credsPath = "./sdk.json"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, err := createAPIClient(ctx)
	if err != nil {
		panic(err)
	}

	// Let's creata a new namespace under the current namespace.
	ns := gaia.NewNamespace()
	ns.Name = "test"

	subctx, subcancel := context.WithTimeout(ctx, 30*time.Second)
	defer subcancel()

	// Create a manipulate context. Note, that we use the subcontext.
	// In case of failures, the manipulate client will do retries
	// until the expiration of the context. If the error is not a
	// timeout or communication error it will return immediately.
	mctx := manipulate.NewContext(subctx)
	if err := m.Create(mctx, ns); err != nil {
		panic(fmt.Sprintf("unable to create namespace: %s", err))
	}

	// Note that the namespace name has been updated with the fully
	// qualified path, even though we provided a relative namespace above.
	fmt.Println("Succesfully created namespace:", ns.Name)

	// Let's create a subscriber for the created namespace. It will
	// listen to policy and external network events.
	s := createSubscriber(ctx, m, ns.Name)

	fmt.Printf("Listening for events in namespace: %s ... \n", ns.Name)
	// We will start a loop now and listen to all events in the namespace
	// and just print them.
	for {
		select {
		case evt := <-s.Events():
			eventHandler(evt)

		case err := <-s.Errors():
			zap.L().Error("Received error from the push channel", zap.Error(err))

		case status := <-s.Status():

			switch status {

			case manipulate.SubscriberStatusDisconnection:
				zap.L().Warn("Upstream event channel interrupted. Reconnecting...")

			case manipulate.SubscriberStatusInitialConnection:
				zap.L().Info("Upstream event channel connected")

			case manipulate.SubscriberStatusReconnection:
				zap.L().Info("Upstream event channel restored")

			case manipulate.SubscriberStatusFinalDisconnection:
				return
			}

		case <-ctx.Done():
			return
		}
	}

}

// creteAPIClient creates a new API client. It uses appcreds
// and configures auto-renewal of the tokens used by the manipulator.
// Note that the API endpoint and root namespace are derived from
// the appcreds automatically.
func createAPIClient(ctx context.Context) (manipulate.Manipulator, error) {
	// Read the appcreds from a local file. The appcreds
	// have been created through apoctl, UI, or the API.
	data, err := ioutil.ReadFile(credsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credential file: %s", err)
	}

	// Unmarshall the data in the data structure.
	appCred, tlsConfig, err := midgardclient.ParseCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credential: %s", err)
	}

	return maniphttp.New(
		ctx,
		appCred.APIURL,
		maniphttp.OptionNamespace(appCred.Namespace),
		maniphttp.OptionTLSConfig(tlsConfig),
		maniphttp.OptionTokenManager(
			tokenmanager.NewX509TokenManager(
				appCred.APIURL,
				24*time.Hour,
				tlsConfig,
			),
		),
	)
}

// createSubscriber creates a new API subscriber. Note that the subscriber
// is created listening on events in the namespace of the manipulator.
func createSubscriber(ctx context.Context, m manipulate.Manipulator, namespace string) manipulate.Subscriber {
	filter := elemental.NewPushFilter()
	filter.FilterIdentity(gaia.NetworkAccessPolicyIdentity.Name)
	filter.FilterIdentity(gaia.ExternalNetworkIdentity.Name)

	s := maniphttp.NewSubscriber(
		m,
		maniphttp.SubscriberOptionRecursive(true),      // get events from children as well.
		maniphttp.SubscriberOptionNamespace(namespace), //start at namespace
	)
	s.Start(ctx, filter)

	return s
}

// eventHandler is a basic handler that just prints the events.
func eventHandler(evt *elemental.Event) {

	switch evt.Identity {

	case gaia.ExternalNetworkIdentity.Name:

		e := gaia.NewExternalNetwork()
		if err := evt.Decode(&e); err != nil {
			zap.L().Error("Failed to decode event", zap.Reflect("event", evt))
		}

		fmt.Printf("External network name: %s type %s\n", e.Name, evt.Type)

	case gaia.NetworkAccessPolicyIdentity.Name:

		p := gaia.NewNetworkAccessPolicy()
		if err := evt.Decode(&p); err != nil {
			zap.L().Error("Failed to decode event", zap.Reflect("event", evt))
		}

		fmt.Printf("Policy name: %s type %s\n", p.Name, evt.Type)

	default:
		zap.L().Error("Received event that was not subscribed", zap.Reflect("event", evt))
	}
}
