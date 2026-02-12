type WeatherRequest = {
  city: string;
};

/**
 * Handles weather requests for different cities
 *
 * @param req - The incoming request containing the city name
 * @returns A JSON response with weather data for the requested city
 */
export async function POST(req: Request) {
  const body: WeatherRequest = await req.json();
  const city = body.city.toLowerCase();

  const weatherData: Record<string, { temperature: number; condition: string; humidity: number; wind: string }> = {
    "new york": { temperature: 45, condition: "Partly Cloudy", humidity: 62, wind: "12 mph NW" },
    "london": { temperature: 48, condition: "Overcast", humidity: 78, wind: "8 mph SW" },
    "tokyo": { temperature: 55, condition: "Clear", humidity: 45, wind: "5 mph E" },
    "sydney": { temperature: 77, condition: "Sunny", humidity: 55, wind: "10 mph SE" },
    "dubai": { temperature: 86, condition: "Sunny", humidity: 40, wind: "7 mph N" },
  };

  const weather = weatherData[city];

  if (weather) {
    return Response.json({ city: body.city, ...weather });
  }

  return Response.json({ city: body.city, error: "City not found" });
}
