import { useQuery } from "@tanstack/react-query";
import { Game } from "@/types";

export function useGames() {
  return useQuery({
    queryKey: ["games"],
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 30000,
    queryFn: async () => {
      const res = await fetch("/api/matches", { credentials: "include" });
      if (!res.ok) {
        console.error("Failed to fetch matches, status:", res.status);
        throw new Error("Failed to fetch games");
      }
      const data = await res.json();
      
      if (data.success && data.matches) {
        return data.matches.map((match: any) => ({
          id: match.id || match.api_id,
          api_id: match.api_id,
          home_team: match.home_team,
          away_team: match.away_team,
          commence_time: match.commence_time,
          home_odds: match.home_odds?.toString() || "1.00",
          draw_odds: match.draw_odds?.toString() || "1.00",
          away_odds: match.away_odds?.toString() || "1.00",
          completed: match.completed || false,
          calculated: match.calculated || false,
          result: match.result || null,
          home_score: match.home_score || null,
          away_score: match.away_score || null,
          created_at: match.created_at || new Date().toISOString(),
          updated_at: match.updated_at || new Date().toISOString(),
        })) as Game[];
      }
      return [] as Game[];
    },
  });
}
