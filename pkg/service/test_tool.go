package service

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

var testTool = mcp.NewTool("weather-tool",
	mcp.WithDescription("A tool for getting weather information"),
	mcp.WithNumber("lat", mcp.Required(), mcp.Description("Latitude of the location")),
	mcp.WithNumber("lon", mcp.Required(), mcp.Description("Longitude of the location")),
	mcp.WithDescription("A tool for getting weather information"),
	mcp.WithOutputSchema[map[string]any](),
)

func (s *Service) testToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Handle the tool request
	log.Println("Tool request with params:", request.Params)
	lat := request.GetFloat("lat", 0)
	lon := request.GetFloat("lon", 0)

	res, err := s.hugr.Query(ctx, `
			query ($lat: Float!, $lon: Float!){
				function{
					owm{
						current_weather(lat: $lat, lon: $lon){
							name
							base
							main{
							temp
							temp_min
							temp_max
							feels_like
							humidity
							pressure
							}
							weather{
							id
							name
							description
							}
						}
					}
				}
			}
		`, map[string]any{
		"lat": lat,
		"lon": lon,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("request error", err), nil
	}
	defer res.Close()

	if len(res.Errors) > 0 {
		return mcp.NewToolResultErrorFromErr("request error", err), nil
	}

	var weather any
	if err := res.ScanData("function.owm.current_weather", &weather); err != nil {
		return mcp.NewToolResultErrorFromErr("read result error", err), nil
	}

	return mcp.NewToolResultStructuredOnly(weather), nil
}
