# Teams

!!! attention "New webhook URL format only"
    Shoutrrr now only supports the new Teams webhook URL format with organization domain.
    You must specify your organization domain using:
    ```text
    ?host=example.webhook.office.com
    ```
    Where `example` is your organization short name.
    
    Legacy webhook formats are no longer supported.

## URL Format

```
teams://group@tenant/altId/groupOwner/extraId?host=organization.webhook.office.com
```

Where:
- `group`: The first UUID component from the webhook URL
- `tenant`: The second UUID component from the webhook URL
- `altId`: The third component (hex string) from the webhook URL
- `groupOwner`: The fourth UUID component from the webhook URL
- `extraId`: The fifth component at the end of the webhook URL
- `organization`: Your organization name for the webhook domain

--8<-- "docs/services/teams/config.md"

## Setting up a webhook

To be able to use the Microsoft Teams notification service, you first need to set up a custom webhook.
Instructions on how to do this can be found in [this guide](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using#setting-up-a-custom-incoming-webhook)

## Extracting the token

The token is extracted from your webhook URL:

<pre><code>https://<b>&lt;organization&gt;</b>.webhook.office.com/webhookb2/<b>&lt;group&gt;</b>@<b>&lt;tenant&gt;</b>/IncomingWebhook/<b>&lt;altId&gt;</b>/<b>&lt;groupOwner&gt;</b>/<b>&lt;extraId&gt;</b></code></pre>

!!! note "Important components"
    All parts of the webhook URL are required:
    
    - `organization`: Your organization name
    - `group`: The first UUID component
    - `tenant`: The second UUID component
    - `altId`: The third component (hex string)
    - `groupOwner`: The fourth UUID component
    - `extraId`: The additional component at the end

## Example

```
# Original webhook URL:
https://contoso.webhook.office.com/webhookb2/11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc/IncomingWebhook/33333333012222222222333333333344/44444444-4444-4444-8444-cccccccccccc/V2ESyij_gAljSoUQHvZoZYzlpAoAXExyOl26dlf1xHEx05

# Shoutrrr URL:
teams://11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc/33333333012222222222333333333344/44444444-4444-4444-8444-cccccccccccc/V2ESyij_gAljSoUQHvZoZYzlpAoAXExyOl26dlf1xHEx05?host=contoso.webhook.office.com
```
