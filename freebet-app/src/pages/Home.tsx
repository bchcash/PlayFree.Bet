import { Link } from "wouter";
import { Navigation } from "@/components/Navigation";
import { Footer } from "@/components/Footer";
import { GamesTable } from "@/components/GamesTable";
import { HelmetManager } from "@/components/HelmetManager";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowRight, Star } from "lucide-react";
import React, { useState, useEffect, useMemo } from "react";
import { useGames } from "@/hooks/use-games";
import { useIsMobile } from "@/hooks/use-mobile";
import { cn } from "@/lib/utils";


const promos = [
  { label: "Telegram", text: "Join our Telegram community for daily tips and updates!", url: "https://t.me/freebet_guru" },
  { label: "X.com", text: "Follow us on X for real-time match discussions!", url: "https://x.com/freebet_guru" },
];

export default function Home() {
  const [promoIndex, setPromoIndex] = useState(0);
  const [showFavorites, setShowFavorites] = useState(false);
  const { data: games } = useGames();
  const isMobile = useIsMobile();

  const filteredGames = useMemo(() => {
    if (!games || !Array.isArray(games)) return [];
    if (!showFavorites) return games;
    try {
      const favoritesString = localStorage.getItem('favorites');
      const favorites = favoritesString ? JSON.parse(favoritesString) : [];
      return games.filter(g => Array.isArray(favorites) && favorites.includes(g.id));
    } catch (e) {
      console.error("Error parsing favorites:", e);
      return games;
    }
  }, [games, showFavorites]);


  useEffect(() => {
    const timer = setInterval(() => {
      setPromoIndex((prev) => (prev + 1) % promos.length);
    }, 5000);
    return () => clearInterval(timer);
  }, []);

  return (
    <>
      <HelmetManager
        title="FreeBet Guru - Virtual Football Betting Platform"
        description="Experience the thrill of virtual football betting! Place free bets on Premier League matches, track live odds, and compete with other players on our leaderboard."
        keywords="virtual betting, free bets, football betting, odds, bookmaker, sports betting, premier league, epl betting, free football bets, soccer betting, virtual bookmaker"
      />
      <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
      <Navigation />
      
      <main className="flex-1 py-12 md:py-20">
        <div className="container mx-auto px-4 max-w-6xl">
          
          {/* Hero Section */}
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center mb-16"
          >
            <span className="inline-block px-3 py-1 rounded-full bg-accent/20 text-xs tracking-wider uppercase mb-4">
              2025/26 Season
            </span>
              <h1 className="text-4xl md:text-6xl font-display font-bold tracking-tight text-foreground mb-4 hover:text-primary transition-colors cursor-pointer">
                Virtual Betting
              </h1>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto mb-8">
              Play for fun, not for money. Our platform uses virtual currency to help you enjoy the game without financial risks and fight gambling addiction.
            </p>

            {/* Promo Slider - Hidden on mobile devices */}
            {!isMobile && (
              <div className="max-w-3xl mx-auto mb-12">
                <a 
                  href={promos[promoIndex].url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="block group"
                >
                  <div className="bg-muted/30 border border-border/50 rounded-xl px-4 py-3 flex items-center gap-4 overflow-hidden h-14 relative hover:border-primary/50 transition-colors">
                    <div className="bg-primary/10 text-primary text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider whitespace-nowrap z-10">
                      {promos[promoIndex].label}
                    </div>
                    <div className="flex-1 relative h-full flex items-center overflow-hidden">
                      <AnimatePresence mode="wait">
                        <motion.div
                          key={promoIndex}
                          initial={{ opacity: 0, y: 20 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{ opacity: 0, y: -20 }}
                          transition={{ duration: 0.5 }}
                          className="text-sm font-medium text-muted-foreground absolute inset-0 flex items-center truncate group-hover:text-foreground transition-colors"
                        >
                          {promos[promoIndex].text}
                        </motion.div>
                      </AnimatePresence>
                    </div>
                    <div className="flex gap-1.5 z-10">
                      {promos.map((_, i) => (
                        <div 
                          key={i} 
                          className={cn(
                            "size-1.5 rounded-full transition-all duration-300",
                            i === promoIndex ? "bg-primary w-4" : "bg-border"
                          )}
                        ></div>
                      ))}
                    </div>
                  </div>
                </a>
              </div>
            )}
          </motion.div>


          {/* Matches Section */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <div className="flex items-center justify-between mb-8">
              <h2 className="text-2xl font-display font-bold flex items-center gap-3">
                Bet Now
                <span className="relative flex h-3 w-3">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
                </span>
              </h2>
              
              <div className="flex items-center justify-end">
                <div
                  className={cn(
                    "flex items-center gap-2 text-sm font-bold transition-all cursor-pointer group",
                    showFavorites ? "text-accent" : "text-primary hover:gap-3"
                  )}
                  onClick={() => setShowFavorites(!showFavorites)}
                >
                  {showFavorites ? "All Matches" : "Favorites"}
                  <ArrowRight className="size-4 group-hover:translate-x-1 transition-transform" />
                </div>
              </div>
            </div>

            {showFavorites && filteredGames.length === 0 ? (
              <div className="p-12 text-center rounded-xl border border-dashed bg-muted/30">
                <Star className="mx-auto size-10 text-muted-foreground mb-3" />
                <h3 className="font-semibold text-lg">No Favorite Matches</h3>
                <p className="text-muted-foreground">Add matches to your favorites to see them here.</p>
              </div>
            ) : (
              <GamesTable favoritesOnly={showFavorites} />
            )}
          </motion.div>

        </div>
      </main>

        <Footer />
      </div>
    </>
  );
}
