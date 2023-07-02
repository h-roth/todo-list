What can we identify about the requests from just the logs?

```
{"time":"2023-07-02T20:44:31.488055+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/ping","http.route":"/ping","http.status":200,"http.request_duration":"28.667µs"}
{"time":"2023-07-02T20:44:47.120964+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists","http.route":"/api/v1/lists","http.status":200,"http.request_duration":"15.292800208s"}
{"time":"2023-07-02T20:44:47.344213+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists/447","http.route":"/api/v1/lists/:list_id","http.status":200,"http.request_duration":"4.738ms"}
{"time":"2023-07-02T20:44:47.521777+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists/447/items","http.route":"/api/v1/lists/:list_id/items","http.status":200,"http.request_duration":"3.749375ms"}
```

Our request to lists is slow!

[ ] 📝 Open backend/cmd/root.go, and edit the Run function:

func Run(ctx context.Context, cfg *Config, logger*slog.Logger, stdout, stderr io.Writer) error {
    ...

    // Setup Tracing: Uncomment this block
    //shutdown, err := setupTracing(ctx, logger)
    //if err != nil {
    // return err
    //}
    //
    //defer shutdown()
    
    ...
You should find the block above. Uncomment the code to enable the tracing provider.

🗣Discuss

What is happening in the setupTracing function?

[ ] 📝 Open backend/todo/routes/router.go, and edit the handler function:

func (m *mux) handler(method, route string, h http.Handler) {
    w := requestLog(h, m.logger, route)
    // Instrument HTTP Handlers: Uncomment the line below
    //w = otelhttp.NewHandler(w, method+" "+route)

    ...
You should find the block above. Uncomment the line containing otelhttp.NewHandler

🗣Discuss

How does this impact the handling of HTTP requests?
Why does this need to be applied here

Why are these requests taking so long?

n+1 problem, for each call to lists API, we make anothher call to the database for every list!

Several solutions, depending on the usecase we could:

* Perform a join in the database as one query
* Split the two database queries but perform one, then the next to get all lists
Essentially anything else that doesnt' involve one repo query calling another.
