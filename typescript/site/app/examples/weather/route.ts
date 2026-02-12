/**
 * Returns weather data for major cities worldwide.
 * Protected by x402 payment — accessed via GET after payment verification.
 */
export async function GET() {
  const weatherData = {
    "New York": { temperature: 45, condition: "Partly Cloudy", humidity: 62, wind: "12 mph NW" },
    "London": { temperature: 48, condition: "Overcast", humidity: 78, wind: "8 mph SW" },
    "Tokyo": { temperature: 55, condition: "Clear", humidity: 45, wind: "5 mph E" },
    "Sydney": { temperature: 77, condition: "Sunny", humidity: 55, wind: "10 mph SE" },
    "Dubai": { temperature: 86, condition: "Sunny", humidity: 40, wind: "7 mph N" },
  };

  return Response.json({ forecast: weatherData });
}
