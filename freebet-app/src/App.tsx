import { Suspense, lazy, useEffect } from "react";
import { Switch, Route } from "wouter";
import { HelmetProvider } from 'react-helmet-async';
import { queryClient } from "./lib/queryClient";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";

// Lazy loading page components for code splitting
const Home = lazy(() => import("@/pages/Home"));
const Leaderboard = lazy(() => import("@/pages/Leaderboard"));
const PlayerBets = lazy(() => import("@/pages/PlayerBets"));
const NotFound = lazy(() => import("@/pages/not-found"));

// Preload critical components
const preloadLeaderboard = () => import("@/pages/Leaderboard");

// Loading component
const LoadingFallback = () => (
  <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
    <div className="flex-1 flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    </div>
  </div>
);

function Router() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Switch>
        <Route path="/" component={Home} />
        <Route path="/leaderboard" component={Leaderboard} />
        <Route path="/player/:nickname" component={PlayerBets} />
        <Route path="/404" component={NotFound} />
        <Route component={NotFound} />
      </Switch>
    </Suspense>
  );
}

function App() {
  // Preload popular pages after app loads
  useEffect(() => {
    const timer = setTimeout(() => {
      preloadLeaderboard();
    }, 2000); // Delay 2 seconds to not block initial loading

    return () => clearTimeout(timer);
  }, []);

  return (
    <HelmetProvider>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          <Toaster />
          <Router />
        </TooltipProvider>
      </QueryClientProvider>
    </HelmetProvider>
  );
}

export default App;
