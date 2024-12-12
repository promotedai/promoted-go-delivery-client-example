# promoted-go-delivery-client-example
Promoted Go Delivery Client Example

## To run

Create a `.env` file with variables.

```
export METRICS_API_ENDPOINT_URL=https://metrics...promoted.ai/log
export METRICS_API_KEY=<metrics api key>

export DELIVERY_API_ENDPOINT_URL=https://delivery...promoted.ai/deliver
export DELIVERY_API_KEY=<delivery api key>
```

Then run

```bash
source .env && go run main.go
```
