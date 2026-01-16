import { format } from "date-fns";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Trophy, DollarSign, Target, TrendingUp, Calendar } from "lucide-react";
import { cn } from "@/lib/utils";
import { Link } from "wouter";

interface PlayerCardProps {
  player: {
    rank: number;
    id: string;
    name: string;
    balance: number;
    bets: number;
    winRate: number;
    avgOdds: number;
    topUps: number;
    pnl: number;
    registeredAt: string;
  };
}

export function PlayerCard({ player }: PlayerCardProps) {
  const getRankDisplay = (rank: number) => {
    if (rank === 1) return "🥇";
    if (rank === 2) return "🥈";
    if (rank === 3) return "🥉";
    return rank;
  };

  return (
    <Link href={`/player/${encodeURIComponent(player.name)}`}>
      <Card className="p-6 hover:shadow-lg transition-all duration-300 border-border/50 bg-gradient-to-br from-card to-card/80 cursor-pointer hover:border-primary/50">
        <div className="flex flex-col space-y-4">
          {/* Header with rank and name */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-primary/10 text-primary font-black text-xl">
                {getRankDisplay(player.rank)}
              </div>
              <div>
                <div className="font-bold text-lg text-foreground">{player.name}</div>
                <div className="flex items-center gap-1 text-sm text-muted-foreground">
                  <Calendar className="size-3" />
                  Joined {format(new Date(player.registeredAt), "MMM yyyy")}
                </div>
              </div>
            </div>
            <Badge className="font-bold bg-primary text-primary-foreground px-3 py-1 rounded-full flex items-center gap-1">
              <Target className="size-3" />
              {player.winRate}%
            </Badge>
          </div>

          {/* Balance and PnL */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-muted/30 rounded-xl p-4 text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <DollarSign className="size-4 text-muted-foreground" />
                <span className="text-xs uppercase text-muted-foreground font-medium">Balance</span>
              </div>
              <div className="font-bold text-lg text-foreground">{player.balance.toLocaleString()}</div>
            </div>
            <div className="bg-muted/30 rounded-xl p-4 text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <TrendingUp className="size-4 text-muted-foreground" />
                <span className="text-xs uppercase text-muted-foreground font-medium">PnL</span>
              </div>
              <div className={cn(
                "font-bold text-lg",
                player.pnl >= 0 ? 'text-green-600 dark:text-green-500' : 'text-red-600 dark:text-red-500'
              )}>
                {player.pnl >= 0 ? `+${player.pnl.toLocaleString()}` : player.pnl.toLocaleString()}
              </div>
            </div>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-4 pt-4 border-t border-border/50">
            <div className="text-center">
              <div className="text-xs uppercase text-muted-foreground font-medium mb-1">Bets</div>
              <div className="font-medium text-foreground">{player.bets}</div>
            </div>
            <div className="text-center">
              <div className="text-xs uppercase text-muted-foreground font-medium mb-1">Avg Odds</div>
              <div className="font-medium text-foreground">{player.avgOdds}</div>
            </div>
            <div className="text-center">
              <div className="text-xs uppercase text-muted-foreground font-medium mb-1">Top-ups</div>
              <div className="font-medium text-foreground">{player.topUps}</div>
            </div>
          </div>
        </div>
      </Card>
    </Link>
  );
}
