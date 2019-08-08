# Aporeto SDK Documentation

This repo contains a simple example for using the Aporeto client SDK.

## Description

The example application will create a namespace and the initialize an
event subscriber and listen on events for the created namespace. It will
listen for external networks and network access policy events only.
The example demonstrates how the API client can be instantiated and how
a subscriber can be defined that listens to specific events.

## Related Libraries

The example uses the Aporeto SKD that is mainly comprised of the following libraries:

1. `manipulate`: is a generic client that can handle API CRUD operations for `elemental` objects. These objects model all the Aporeto API. (https://github.com/aporeto-inc/manipulate)

2. `elemental`: is a generic library of models that represent API resources. (https://github.com/aporeto-inc/elemental)

3. `gaia`: is the actual definition of the Aporeto API. (https://github.com/aporeto-inc/gaia)

4. `midgard-lib`: is a library of helpers that deals with authentication against the platform. (https://github.com/aporeto-inc/midgard-lib)

## Installation

1. Create appcreds either in the Aporeto UI or using the apoctl utility. Store the appcreds file in the `./example` directory as sdk.json.

2. Navigate to example and run it:

```bash
go run main.go
```

Always use a dependency manager to fix library versions to the current release of your installation.
