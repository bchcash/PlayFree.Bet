import { useState } from "react";
import { Game } from "@/types";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { format, isToday, isTomorrow } from "date-fns";
import { Star, Clock } from "lucide-react";
import { BettingModal } from "@/components/BettingModal";
import { useUser } from "@/hooks/use-user";
import { useToast } from "@/hooks/use-toast";
import { getLogoWithFallback } from "@/lib/logos";
import { OptimizedImage } from "@/components/ui/optimized-image";
import { cn } from "@/lib/utils";

interface BettingCardProps {
  game: Game;
}

export function BettingCard({ game }: BettingCardProps) {
  const { toast } = useToast();
  const { data: user, isLoading: userLoading } = useUser();
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedBet, setSelectedBet] = useState<{
    betType: "home" | "draw" | "away";
    odds: number;
  } | null>(null);

  const [isFavorite, setIsFavorite] = useState(() => {
    const favorites = JSON.parse(localStorage.getItem('favorites') || '[]');
    return favorites.includes(game.id);
  });

  const toggleFavorite = (e: React.MouseEvent) => {
    e.stopPropagation();
    const favorites = JSON.parse(localStorage.getItem('favorites') || '[]');
    let newFavorites;
    if (isFavorite) {
      newFavorites = favorites.filter((id: string) => id !== game.id);
    } else {
      newFavorites = [...favorites, game.id];
    }
    localStorage.setItem('favorites', JSON.stringify(newFavorites));
    setIsFavorite(!isFavorite);
    
    // Dispatch custom event to notify other components (like Home)
    window.dispatchEvent(new Event('favoritesUpdated'));
  };

  const getLogo = (teamName: string) => {
    return getLogoWithFallback(teamName) || null;
  };

  const handleOddsClick = (betType: "home" | "draw" | "away") => {
    if (userLoading) {
      return;
    }
    
    if (user === null) {
      toast({
        title: "Login Required",
        description: "Please login to place bets.",
        variant: "destructive",
      });
      return;
    }

    const odds =
      betType === "home"
        ? parseFloat(String(game.home_odds))
        : betType === "draw"
          ? parseFloat(String(game.draw_odds))
          : parseFloat(String(game.away_odds));

    setSelectedBet({ betType, odds });
    setModalOpen(true);
  };

  const handleModalClose = () => {
    setModalOpen(false);
    setSelectedBet(null);
  };

  return (
    <>
      <Card className="p-5 bg-[#F4F4F4] dark:bg-[#1A1A1A] hover:shadow-md transition-all duration-300 border border-[#E0E0E0] dark:border-[#333333] rounded-[24px] relative overflow-hidden h-full flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2">
              <div className="bg-primary text-primary-foreground text-[10px] px-2 py-0.5 rounded-[4px] font-bold uppercase tracking-tight">
                Football
              </div>
              <span className="text-[11px] text-muted-foreground font-light tracking-tight truncate max-w-[150px]">
                England. Premier League
              </span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className={cn(
                "h-6 w-6 transition-colors",
                isFavorite ? "text-yellow-500 fill-yellow-500" : "text-[#9E9E9E] hover:text-primary"
              )}
              onClick={toggleFavorite}
            >
              <Star className={cn("h-4.5 w-4.5", isFavorite && "fill-current")} />
            </Button>
          </div>

          <div className="flex items-center justify-between gap-2 mb-8 px-2 min-h-[100px]">
            <div className="flex flex-col items-center flex-1">
              <div className="w-14 h-14 flex items-center justify-center mb-3">
                {getLogo(game.home_team) ? (
                  <OptimizedImage
                    src={getLogo(game.home_team)!}
                    alt={`${game.home_team} football club logo - team emblem`}
                    fallbackSrc={getLogo(game.home_team)!.replace('.webp', '.png')}
                    className="w-12 h-12 object-contain"
                  />
                ) : (
                  <div className="w-full h-full rounded-full bg-white dark:bg-[#2A2A2A] flex items-center justify-center text-sm font-medium text-[#2D3139] dark:text-white shadow-sm border border-[#E0E0E0] dark:border-[#444444]" aria-label={`${game.home_team} team initial`}>
                    {game.home_team[0]}
                  </div>
                )}
              </div>
              <span className="text-[13px] font-medium text-[#2D3139] dark:text-white text-center leading-[1.2] line-clamp-2 h-8 flex items-center tracking-tighter uppercase font-mono">
                {game.home_team}
              </span>
            </div>

            <div className="flex flex-col items-center justify-center min-w-[90px]">
              <div className="flex items-center gap-1 mb-1">
                <Clock className="size-4 text-[#9E9E9E]" />
                <span className="text-[22px] font-medium text-[#2D3139] dark:text-white leading-none">
                  {format(new Date(game.commence_time), "HH:mm")}
                </span>
              </div>
              <span className="text-[12px] text-[#9E9E9E] font-normal">
                {isToday(new Date(game.commence_time))
                  ? "Today"
                  : isTomorrow(new Date(game.commence_time))
                    ? "Tomorrow"
                    : format(new Date(game.commence_time), "dd MMM")}
              </span>
            </div>

            <div className="flex flex-col items-center flex-1">
              <div className="w-14 h-14 flex items-center justify-center mb-3">
                {getLogo(game.away_team) ? (
                  <OptimizedImage
                    src={getLogo(game.away_team)!}
                    alt={`${game.away_team} football club logo - team emblem`}
                    fallbackSrc={getLogo(game.away_team)!.replace('.webp', '.png')}
                    className="w-12 h-12 object-contain"
                  />
                ) : (
                  <div className="w-full h-full rounded-full bg-white dark:bg-[#2A2A2A] flex items-center justify-center text-sm font-medium text-[#2D3139] dark:text-white shadow-sm border border-[#E0E0E0] dark:border-[#444444]" aria-label={`${game.away_team} team initial`}>
                    {game.away_team[0]}
                  </div>
                )}
              </div>
              <span className="text-[13px] font-medium text-[#2D3139] dark:text-white text-center leading-[1.2] line-clamp-2 h-8 flex items-center tracking-tighter uppercase font-mono">
                {game.away_team}
              </span>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2 px-1">
          <Button
            variant="ghost"
            className="h-[52px] flex items-center justify-between px-4 bg-white dark:bg-[#2A2A2A] hover:bg-green-100 dark:hover:bg-green-900/30 hover:border-green-300 dark:hover:border-green-600 hover:shadow-lg hover:shadow-green-200/50 dark:hover:shadow-green-900/30 transition-all duration-300 rounded-[10px] border-2 border-[#E0E0E0] dark:border-[#444444] shadow-sm group hover:scale-[1.02] active:scale-[0.98]"
            onClick={() => handleOddsClick("home")}
          >
            <span className="text-[11px] font-medium text-[#9E9E9E] group-hover:text-green-700 dark:group-hover:text-green-300 transition-colors">
              1
            </span>
            <span className="text-[15px] font-bold text-[#2D3139] dark:text-white group-hover:text-green-800 dark:group-hover:text-green-100 group-hover:scale-110 transition-all">
              {game.home_odds}
            </span>
          </Button>
          <Button
            variant="ghost"
            className="h-[52px] flex items-center justify-between px-4 bg-white dark:bg-[#2A2A2A] hover:bg-green-100 dark:hover:bg-green-900/30 hover:border-green-300 dark:hover:border-green-600 hover:shadow-lg hover:shadow-green-200/50 dark:hover:shadow-green-900/30 transition-all duration-300 rounded-[10px] border-2 border-[#E0E0E0] dark:border-[#444444] shadow-sm group hover:scale-[1.02] active:scale-[0.98]"
            onClick={() => handleOddsClick("draw")}
          >
            <span className="text-[11px] font-medium text-[#9E9E9E] group-hover:text-green-700 dark:group-hover:text-green-300 transition-colors">
              X
            </span>
            <span className="text-[15px] font-bold text-[#2D3139] dark:text-white group-hover:text-green-800 dark:group-hover:text-green-100 group-hover:scale-110 transition-all">
              {game.draw_odds}
            </span>
          </Button>
          <Button
            variant="ghost"
            className="h-[52px] flex items-center justify-between px-4 bg-white dark:bg-[#2A2A2A] hover:bg-green-100 dark:hover:bg-green-900/30 hover:border-green-300 dark:hover:border-green-600 hover:shadow-lg hover:shadow-green-200/50 dark:hover:shadow-green-900/30 transition-all duration-300 rounded-[10px] border-2 border-[#E0E0E0] dark:border-[#444444] shadow-sm group hover:scale-[1.02] active:scale-[0.98]"
            onClick={() => handleOddsClick("away")}
          >
            <span className="text-[11px] font-medium text-[#9E9E9E] group-hover:text-green-700 dark:group-hover:text-green-300 transition-colors">
              2
            </span>
            <span className="text-[15px] font-bold text-[#2D3139] dark:text-white group-hover:text-green-800 dark:group-hover:text-green-100 group-hover:scale-110 transition-all">
              {game.away_odds}
            </span>
          </Button>
        </div>
      </Card>

      {selectedBet && (
        <BettingModal
          open={modalOpen}
          onClose={handleModalClose}
          matchId={game.id}
          homeTeam={game.home_team}
          awayTeam={game.away_team}
          betType={selectedBet.betType}
          odds={selectedBet.odds}
          userBalance={user?.money || 0}
        />
      )}
    </>
  );
}
