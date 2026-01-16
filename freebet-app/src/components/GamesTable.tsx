import { useGames } from "@/hooks/use-games";
import { Calendar, ChevronLeft, ChevronRight } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { BettingCard } from "./BettingCard";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useState, useMemo } from "react";

export function GamesTable({ favoritesOnly = false }: { favoritesOnly?: boolean }) {
  const { data: games, isLoading, isError } = useGames();
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 9;

  const filteredGames = useMemo(() => {
    if (!games) return [];
    if (!favoritesOnly) return games;
    const favorites = JSON.parse(localStorage.getItem('favorites') || '[]');
    return games.filter(g => favorites.includes(g.id));
  }, [games, favoritesOnly]);

  const totalPages = useMemo(() => {
    return Math.ceil(filteredGames.length / itemsPerPage);
  }, [filteredGames]);

  const paginatedGames = useMemo(() => {
    return filteredGames.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
  }, [filteredGames, currentPage]);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((i) => (
          <Skeleton key={i} className="h-[220px] w-full rounded-2xl" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8 text-center rounded-xl bg-destructive/10 text-destructive border border-destructive/20">
        <p className="font-semibold">Unable to load match schedule.</p>
        <p className="text-sm opacity-80">Please try again later.</p>
      </div>
    );
  }

  if (!games || games.length === 0) {
    return (
      <div className="p-12 text-center rounded-xl border border-dashed bg-muted/30">
        <Calendar className="mx-auto size-10 text-muted-foreground mb-3" />
        <h3 className="font-semibold text-lg">No Upcoming Matches</h3>
        <p className="text-muted-foreground">The fixture list is currently empty.</p>
      </div>
    );
  }

  return (
    <div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {paginatedGames.map((game) => (
          <BettingCard key={game.id} game={game} />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8 py-4">
          <Button 
            variant="ghost" 
            size="icon" 
            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
            disabled={currentPage === 1}
            className="rounded-md hover:bg-transparent transition-all group"
          >
            <ChevronLeft className="size-4 group-hover:text-primary group-hover:-translate-x-0.5 transition-transform" />
          </Button>
          <div className="flex items-center gap-2 px-4">
            {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
              <Button
                key={page}
                variant="ghost"
                size="sm"
                onClick={() => setCurrentPage(page)}
                className={cn(
                  "w-10 h-10 p-0 rounded-md transition-all font-bold text-sm hover:bg-transparent",
                  currentPage === page 
                    ? "text-primary font-bold" 
                    : "text-muted-foreground hover:text-primary"
                )}
              >
                {page}
              </Button>
            ))}
          </div>
          <Button 
            variant="ghost" 
            size="icon" 
            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages}
            className="rounded-md hover:bg-transparent transition-all group"
          >
            <ChevronRight className="size-4 group-hover:text-primary group-hover:translate-x-0.5 transition-transform" />
          </Button>
        </div>
      )}
    </div>
  );
}
