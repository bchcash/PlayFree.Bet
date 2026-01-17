import { Link, useLocation } from "wouter";
import { cn } from "@/lib/utils";
import {
  Trophy,
  CalendarDays,
  Users,
  LogOut,
  User as UserIcon,
  Settings,
  ShieldCheck,
  Wallet,
  BarChart3,
  TrendingUp,
  Lock,
  PlusCircle,
  AlertCircle,
  LogIn,
  UserPlus,
  Menu,
  X,
  Heart,
  Copy
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { useState, useEffect } from "react";
import { useToast } from "@/hooks/use-toast";
import { useQueryClient } from "@tanstack/react-query";
import { getAccessToken, setAccessToken, getRefreshToken, setRefreshToken, removeAccessToken, removeRefreshToken } from "@/hooks/use-user";

interface User {
  id: string;
  email: string;
  nickname: string;
  money: number;
  topup: number;
  last_topup_at?: string | null;
}

import { useUser } from "@/hooks/use-user";

export function Navigation() {
  const [location] = useLocation();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [isSupportModalOpen, setIsSupportModalOpen] = useState(false);
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login');
  const { data: user, isLoading: isUserLoading } = useUser();
  const [authForm, setAuthForm] = useState({ identifier: '', password: '', nickname: '', ageConfirmed: false });
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const pnl = user ? user.money - (user.topup * 10000) : 0;
  const isTopUpDisabled = (user?.money || 0) >= 500;

  // Listen for external events to open auth modal
  useEffect(() => {
    const handleOpenAuthModal = () => {
      setAuthMode('login');
      setIsAuthModalOpen(true);
    };

    window.addEventListener('openAuthModal', handleOpenAuthModal);
    return () => window.removeEventListener('openAuthModal', handleOpenAuthModal);
  }, []);

  const handleTopUp = async () => {
    if ((user?.money || 0) >= 500) {
      toast({
        title: "Top-up not available",
        description: "Your balance must be below 500 to top-up.",
        variant: "destructive",
      });
      return;
    }

    // Check if user has already topped up within last 24 hours
    if (user?.last_topup_at) {
      const lastTopupTime = new Date(user.last_topup_at);
      const now = new Date();
      const timeSinceLastTopup = now.getTime() - lastTopupTime.getTime();
      const hoursSinceLastTopup = timeSinceLastTopup / (1000 * 60 * 60);

      if (hoursSinceLastTopup < 24) {
        const hoursRemaining = Math.floor(24 - hoursSinceLastTopup);
        const minutesRemaining = Math.floor((24 - hoursSinceLastTopup) * 60) % 60;
        toast({
          title: "Top-up not available",
          description: (
            <div>
              <div>You can only top up once per day.</div>
              <div>Please wait {hoursRemaining} hours and {minutesRemaining} minutes.</div>
            </div>
          ),
          variant: "destructive",
        });
        return;
      }
    }

    try {
      const response = await fetch('/api/auth/topup', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (response.ok) {
        queryClient.invalidateQueries({ queryKey: ['user'] });
        toast({
          title: "Top-up successful",
          description: "10,000 has been added to your account.",
        });
      } else {
        const data = await response.json();
        toast({
          title: "Top-up failed",
          description: data.error || "Please try again later.",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Top-up failed",
        description: "Network error. Please try again.",
        variant: "destructive",
      });
    }
  };

  const handleChangePassword = async () => {
    const currentPassword = (document.getElementById('current') as HTMLInputElement)?.value;
    const newPassword = (document.getElementById('new') as HTMLInputElement)?.value;
    const confirmPassword = (document.getElementById('confirm') as HTMLInputElement)?.value;

    if (!currentPassword || !newPassword || !confirmPassword) {
      toast({
        title: "Error",
        description: "Please fill in all fields.",
        variant: "destructive",
      });
      return;
    }

    if (newPassword !== confirmPassword) {
      toast({
        title: "Error",
        description: "New passwords don't match.",
        variant: "destructive",
      });
      return;
    }

    try {
      const response = await fetch('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      });

      if (response.ok) {
        setIsPasswordModalOpen(false);
        // Clear form
        (document.getElementById('current') as HTMLInputElement).value = '';
        (document.getElementById('new') as HTMLInputElement).value = '';
        (document.getElementById('confirm') as HTMLInputElement).value = '';
        toast({
          title: "Success",
          description: "Password changed successfully.",
        });
      } else {
        const data = await response.json();
        toast({
          title: "Error",
          description: data.error || "Failed to change password.",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Network error. Please try again.",
        variant: "destructive",
      });
    }
  };

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;
    return emailRegex.test(email);
  };

  const handleAuth = async () => {
    // Validate email format for registration
    if (authMode === 'register') {
      if (!validateEmail(authForm.identifier)) {
        toast({
          title: "Invalid email",
          variant: "destructive",
          description: "Please enter a valid email address",
        });
        return;
      }
      if (authForm.nickname.length < 3 || authForm.nickname.length > 10) {
        toast({
          title: "Invalid nickname",
          variant: "destructive",
          description: "Nickname must be between 3 and 10 characters",
        });
        return;
      }

      if (!authForm.ageConfirmed) {
        toast({
          title: "Age confirmation required",
          variant: "destructive",
          description: "You must confirm that you are 18 years or older",
        });
        return;
      }
    }

    setIsLoading(true);
    try {
      const endpoint = authMode === 'login' ? '/api/auth/login' : '/api/auth/register';
      const body = authMode === 'login'
        ? { identifier: authForm.identifier, password: authForm.password }
        : { email: authForm.identifier, password: authForm.password, nickname: authForm.nickname, age_confirmed: authForm.ageConfirmed };

      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(body),
      });

      const data = await res.json();

      if (data.success && data.user && data.access_token) {
        // Save JWT tokens
        setAccessToken(data.access_token);
        if (data.refresh_token) {
          setRefreshToken(data.refresh_token);
        }

        setIsAuthModalOpen(false);
        setAuthForm({ identifier: '', password: '', nickname: '', ageConfirmed: false });
        queryClient.invalidateQueries({ queryKey: ['user'] });
        queryClient.invalidateQueries({ queryKey: ['user-bets'] });
        toast({
          title: "Success",
          description: authMode === 'login' ? "Logged in successfully" : "Account created successfully",
        });
      } else {
        toast({
          title: "Error",
          variant: "destructive",
          description: data.error || "Authentication failed",
        });
      }
    } catch (error) {
      toast({
        title: "Error",
        variant: "destructive",
        description: "Network error. Please try again.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleGoogleAuth = () => {
    // Redirect to Google OAuth with current URL as redirect_url
    const currentUrl = window.location.origin + window.location.pathname;
    window.location.href = `/api/auth/google?redirect_url=${encodeURIComponent(currentUrl)}`;
  };

  // Handle Google OAuth callback
  useEffect(() => {
    const handleGoogleCallback = () => {
      const urlParams = new URLSearchParams(window.location.search);
      const accessToken = urlParams.get('access_token');
      const refreshToken = urlParams.get('refresh_token');

      if (accessToken) {
        setAccessToken(accessToken);
        if (refreshToken) {
          setRefreshToken(refreshToken);
        }

        // Clear tokens from URL
        const newUrl = window.location.pathname;
        window.history.replaceState({}, document.title, newUrl);

        // User data will be automatically refreshed due to token change in queryKey
        toast({
          title: "Success",
          description: "Successfully logged in with Google!",
        });
      }
    };

    handleGoogleCallback();
  }, [queryClient, toast]);

  const handleLogout = async () => {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
          'Content-Type': 'application/json',
        },
        credentials: 'include'
      });

      // Clear JWT tokens
      removeAccessToken();
      removeRefreshToken();

      queryClient.setQueryData(['user'], null);
      queryClient.invalidateQueries({ queryKey: ['user'] });
      queryClient.invalidateQueries({ queryKey: ['user-bets'] });
      toast({ title: "Success", description: "Logged out successfully" });
    } catch (error) {
      // Even if logout request fails, clear local tokens
      removeAccessToken();
      removeRefreshToken();
      queryClient.setQueryData(['user'], null);
      queryClient.invalidateQueries({ queryKey: ['user'] });
      queryClient.invalidateQueries({ queryKey: ['user-bets'] });
      toast({ title: "Success", description: "Logged out locally" });
    }
  };

  const navItems = [
    { label: "Matches", href: "/", icon: CalendarDays },
    { label: "Leaderboard", href: "/leaderboard", icon: Trophy },
  ];

  return (
    <>
      <header className="sticky top-0 z-50 w-full border-b bg-background/80 backdrop-blur-md supports-[backdrop-filter]:bg-background/60 h-16">
        <div className="container mx-auto px-4 h-full flex items-center justify-between">
          {/* Mobile Menu Button */}
          <div className="flex items-center lg:hidden">
            <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
              <SheetTrigger asChild>
                <Button 
                  variant="ghost" 
                  size="icon" 
                  className="relative w-10 h-10 rounded-xl hover:bg-primary/10 transition-all"
                >
                  <div className="flex flex-col items-center justify-center gap-1.5">
                    <span className={cn(
                      "block h-0.5 w-5 bg-foreground rounded-full transition-all duration-300",
                      isMobileMenuOpen && "rotate-45 translate-y-2"
                    )} />
                    <span className={cn(
                      "block h-0.5 w-5 bg-foreground rounded-full transition-all duration-300",
                      isMobileMenuOpen && "opacity-0"
                    )} />
                    <span className={cn(
                      "block h-0.5 w-5 bg-foreground rounded-full transition-all duration-300",
                      isMobileMenuOpen && "-rotate-45 -translate-y-2"
                    )} />
                  </div>
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-[280px] p-0 border-r border-border/50 flex flex-col">
                <SheetHeader className="p-6 border-b border-border/50 flex-shrink-0">
                  <SheetTitle className="font-display font-black text-xl tracking-tighter text-foreground italic text-left">
                    FREEBET<span className="text-primary not-italic">GURU</span>
                  </SheetTitle>
                </SheetHeader>
                <nav className="flex flex-col p-4 gap-1 flex-1 overflow-y-auto">
                  {navItems.map((item) => (
                    <SheetClose asChild key={item.href}>
                      <Link
                        href={item.href}
                        className={cn(
                          "px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3",
                          location === item.href
                            ? "bg-primary/10 text-primary font-semibold"
                            : "text-muted-foreground hover:bg-muted hover:text-foreground"
                        )}
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        <item.icon className="size-5" />
                        {item.label}
                      </Link>
                    </SheetClose>
                  ))}

                  {/* User Menu Items for Mobile */}
                  {user && (
                    <>
                      <div className="border-t border-border/50 my-2" />
                      <SheetClose asChild>
                        <Link
                          href={`/player/${user.nickname}`}
                          className="px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3 text-muted-foreground hover:bg-muted hover:text-foreground"
                          onClick={() => setIsMobileMenuOpen(false)}
                        >
                          <TrendingUp className="size-5" />
                          My Bets
                        </Link>
                      </SheetClose>
                      <SheetClose asChild>
                        <button
                          className="w-full px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3 text-muted-foreground hover:bg-muted hover:text-foreground text-left"
                          onClick={() => {
                            handleTopUp();
                            setIsMobileMenuOpen(false);
                          }}
                        >
                          <Wallet className="size-5" />
                          Top Up Balance
                        </button>
                      </SheetClose>
                      <SheetClose asChild>
                        <button
                          className="w-full px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3 text-muted-foreground hover:bg-muted hover:text-foreground text-left"
                          onClick={() => {
                            setIsSupportModalOpen(true);
                            setIsMobileMenuOpen(false);
                          }}
                        >
                          <Heart className="size-5" />
                          Support the project
                        </button>
                      </SheetClose>
                      {/* Only show Change Password for email users */}
                      {user.auth_provider !== 'google' && (
                        <SheetClose asChild>
                          <button
                            className="w-full px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3 text-muted-foreground hover:bg-muted hover:text-foreground text-left"
                            onClick={() => {
                              setIsPasswordModalOpen(true);
                              setIsMobileMenuOpen(false);
                            }}
                          >
                            <Lock className="size-5" />
                            Change Password
                          </button>
                        </SheetClose>
                      )}
                      <div className="border-t border-border/50 my-2" />
                      <SheetClose asChild>
                        <button
                          className="w-full px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 flex items-center gap-3 text-muted-foreground hover:bg-muted hover:text-foreground text-left text-destructive hover:bg-destructive/10 hover:text-destructive"
                          onClick={() => {
                            handleLogout();
                            setIsMobileMenuOpen(false);
                          }}
                        >
                          <LogOut className="size-5" />
                          Log out
                        </button>
                      </SheetClose>
                    </>
                  )}
                </nav>

                {/* Bottom section - always visible */}
                {!user && (
                  <div className="p-4 border-t border-border/50 mt-auto flex-shrink-0">
                    <div className="flex flex-col gap-2">
                      <Button
                        className="w-full rounded-xl font-bold uppercase text-[10px] tracking-widest h-11 bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] flex items-center justify-center"
                        onClick={() => {
                          setAuthMode('login');
                          setIsAuthModalOpen(true);
                          setIsMobileMenuOpen(false);
                        }}
                      >
                        Get Started
                      </Button>
                    </div>
                  </div>
                )}
              </SheetContent>
            </Sheet>
          </div>

          {/* Logo - hidden on mobile */}
          <Link href="/" className="hidden lg:flex items-center gap-2">
            <span className="font-display font-black text-2xl tracking-tighter text-foreground italic">
              FREEBET<span className="text-primary not-italic">GURU</span>
            </span>
          </Link>

          {/* Center Nav */}
          <nav className="hidden lg:flex items-center gap-1">
            {navItems.map((item) => (
              <Link 
                key={item.href} 
                href={item.href}
                className={cn(
                  "px-4 py-2 rounded-full text-sm font-medium transition-all duration-200 flex items-center gap-2",
                  location === item.href 
                    ? "bg-primary/10 text-primary font-semibold" 
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                )}
              >
                <item.icon className="size-4" />
                {item.label}
              </Link>
            ))}
          </nav>

          {/* Right Actions */}
          <div className="flex items-center gap-4">
            {isUserLoading ? (
              <div className="w-24 h-9 bg-muted/50 rounded-full animate-pulse" />
            ) : !user ? (
              <Button 
                variant="default" 
                size="sm" 
                className="font-bold uppercase text-[10px] tracking-widest rounded-full px-6 h-9 shadow-[0_0_20px_rgba(var(--primary),0.3)] hover:shadow-[0_0_25px_rgba(var(--primary),0.5)] transition-all bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary hover:scale-[1.02] active:scale-95 flex items-center justify-center"
                onClick={() => {
                  setAuthMode('login');
                  setIsAuthModalOpen(true);
                }}
              >
                Get Started
              </Button>
            ) : (
              <>
                <div className="flex items-center gap-3 sm:gap-4 mr-2 animate-in fade-in slide-in-from-right-2 duration-500">
                  <div className="flex flex-col items-end">
                    <span className="text-[8px] sm:text-[10px] uppercase font-bold text-muted-foreground leading-none mb-0.5 sm:mb-1">Bal</span>
                    <span className="text-xs sm:text-sm font-black text-foreground tabular-nums">
                      ${user.money.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                    </span>
                  </div>
                  <div className="h-6 sm:h-8 w-px bg-border/50" />
                  <div className="flex flex-col items-end">
                    <span className="text-[8px] sm:text-[10px] uppercase font-bold text-muted-foreground leading-none mb-0.5 sm:mb-1">PnL</span>
                    <span className={cn("text-xs sm:text-sm font-black tabular-nums", pnl >= 0 ? "text-green-500" : "text-red-500")}>
                      {pnl >= 0 ? "+" : "-"}${Math.abs(pnl).toLocaleString(undefined, { maximumFractionDigits: 0 })}
                    </span>
                  </div>
                  <div className="h-6 sm:h-8 w-px bg-border/50" />
                </div>

                <DropdownMenu>
                  <DropdownMenuTrigger className="focus:outline-none">
                    <div className="flex items-center gap-2 px-1 py-1 sm:px-2 sm:py-1.5 rounded-full hover:bg-muted/50 transition-colors border border-transparent hover:border-border cursor-pointer">
                      <Avatar className="size-7 sm:size-8 border-2 border-background shadow-sm">
                        <AvatarFallback>{user.nickname.slice(0, 2).toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="text-xs text-left hidden sm:block mr-1">
                        <p className="font-semibold text-foreground">{user.nickname}</p>
                        <p className="text-muted-foreground">Player</p>
                      </div>
                    </div>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-56 mt-2 rounded-xl border shadow-xl p-2">
                    <DropdownMenuLabel className="px-2 py-1.5 text-xs text-muted-foreground uppercase font-bold tracking-wider">My Account</DropdownMenuLabel>
                    <DropdownMenuItem 
                      className="rounded-lg cursor-pointer focus:bg-primary/10 focus:text-primary"
                      asChild
                    >
                      <Link href={`/player/${user.nickname}`} className="flex items-center w-full">
                        <TrendingUp className="mr-2 size-4" />
                        <span>My Bets</span>
                      </Link>
                    </DropdownMenuItem>
                    {/* Only show Change Password for email users */}
                    {user.auth_provider !== 'google' && (
                      <DropdownMenuItem
                        className="rounded-lg cursor-pointer focus:bg-primary/10 focus:text-primary"
                        onSelect={() => setIsPasswordModalOpen(true)}
                      >
                        <Lock className="mr-2 size-4" />
                        <span>Change Password</span>
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuItem
                      className="rounded-lg cursor-pointer focus:bg-primary/10 focus:text-primary"
                      onSelect={handleTopUp}
                    >
                      <Wallet className="mr-2 size-4" />
                      <span>Top Up Balance</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="rounded-lg cursor-pointer focus:bg-primary/10 focus:text-primary"
                      onClick={() => setIsSupportModalOpen(true)}
                    >
                      <Heart className="mr-2 size-4" />
                      <span>Support the project</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator className="my-1" />
                    <DropdownMenuItem 
                      className="rounded-lg cursor-pointer text-destructive focus:bg-destructive/10 focus:text-destructive"
                      onClick={handleLogout}
                    >
                      <LogOut className="mr-2 size-4" />
                      <span>Log out</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Auth Modal */}
      <Dialog open={isAuthModalOpen} onOpenChange={setIsAuthModalOpen}>
        <DialogContent className="sm:max-w-[425px] rounded-2xl" aria-describedby="auth-description">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold tracking-tight flex items-center gap-2">
              {authMode === 'login' ? 'Login' : 'Register'}
            </DialogTitle>
            <p id="auth-description" className="text-sm text-muted-foreground">
              {authMode === 'login'
                ? 'Access your account to continue betting'
                : 'Create an account to start your betting journey'}
            </p>
          </DialogHeader>
          <div className="grid gap-6 py-4">
            <div className="grid gap-2">
              <Label htmlFor="identifier" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                {authMode === 'login' ? 'Email or Nickname' : 'Email'}
              </Label>
              <Input
                id="identifier"
                type={authMode === 'login' ? 'text' : 'email'}
                placeholder={authMode === 'login' ? 'Enter email or nickname' : 'Enter your email'}
                className="rounded-xl border-border/50 focus:ring-primary/20"
                value={authForm.identifier}
                onChange={(e) => setAuthForm(prev => ({ ...prev, identifier: e.target.value }))}
              />
            </div>
            {authMode === 'register' && (
              <>
                <div className="grid gap-2">
                  <Label htmlFor="nickname" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                    Nickname
                  </Label>
                  <Input
                    id="nickname"
                    type="text"
                    placeholder="Choose a nickname"
                    className="rounded-xl border-border/50 focus:ring-primary/20"
                    value={authForm.nickname}
                    onChange={(e) => setAuthForm(prev => ({ ...prev, nickname: e.target.value }))}
                  />
                </div>
              </>
            )}
            <div className="grid gap-2">
              <Label htmlFor="password" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                Password
              </Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                className="rounded-xl border-border/50 focus:ring-primary/20"
                value={authForm.password}
                onChange={(e) => setAuthForm(prev => ({ ...prev, password: e.target.value }))}
              />
            </div>
            {authMode === 'register' && (
              <div className="grid gap-2">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="age-confirm"
                    checked={authForm.ageConfirmed}
                    onCheckedChange={(checked) =>
                      setAuthForm(prev => ({ ...prev, ageConfirmed: checked as boolean }))
                    }
                  />
                  <Label htmlFor="age-confirm" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                    I confirm that I am 18 years or older
                  </Label>
                </div>
              </div>
            )}

          <DialogFooter>
            <Button
              type="submit"
              className="w-full rounded-xl font-bold uppercase tracking-wider h-11"
              onClick={handleAuth}
              disabled={isLoading}
            >
              {isLoading ? 'Please wait...' : (authMode === 'login' ? 'Sign In' : 'Create Account')}
            </Button>
          </DialogFooter>

              {/* Google OAuth Button */}
              <div className="mt-6">
                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t border-border/50" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase tracking-widest">
                    <span className="bg-background px-2 text-muted-foreground font-light">or continue with</span>
                  </div>
                </div>

                <Button
                  variant="outline"
                  className="w-full rounded-2xl h-12 mt-4 border-border/40 hover:bg-muted/40 transition-all text-sm font-light"
                  onClick={handleGoogleAuth}
                  disabled={isLoading}
                >
                  <svg className="w-4 h-4 mr-2" viewBox="0 0 24 24">
                    <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                    <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                    <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                    <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                  </svg>
                  Continue with Google
                </Button>
              </div>

              <div className="text-center pt-6">
                <button
                  className="group relative inline-flex flex-col items-center gap-1"
                  onClick={() => setAuthMode(authMode === 'login' ? 'register' : 'login')}
                >
                  <span className="text-[10px] text-muted-foreground font-light uppercase tracking-widest">
                    {authMode === 'login' ? "New to the platform?" : "Already verified?"}
                  </span>
                  <span className="text-xs font-light text-primary uppercase tracking-widest group-hover:underline underline-offset-4 decoration-2">
                    {authMode === 'login' ? 'Register Now' : 'Sign In Instead'}
                  </span>
                </button>
              </div>
            </div>
        </DialogContent>
      </Dialog>


      {/* Support Modal */}
      <Dialog open={isSupportModalOpen} onOpenChange={setIsSupportModalOpen}>
        <DialogContent className="sm:max-w-[400px] rounded-2xl" aria-describedby="support-description">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold tracking-tight flex items-center gap-2">
              <Heart className="size-5 text-red-500" />
              Support the project
            </DialogTitle>
            <p id="support-description" className="text-sm text-muted-foreground">
              Help us keep FreeBet Guru running and growing!
            </p>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="text-center">
              <span className="text-sm font-medium text-muted-foreground">USDT TRC-20</span>
            </div>

            <div className="p-4 bg-muted/30 border border-border/50 rounded-xl">
              <div className="text-center mb-4">
                <img
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=TTx1Lr2ZLAKW7CBJMKyxfWXZoEFrjH8edN`}
                  alt="USDT TRC-20 QR Code"
                  className="w-32 h-32 mx-auto rounded-lg border border-border"
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                  Wallet Address
                </label>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-xs font-mono bg-background px-3 py-2 rounded-lg border border-border/50 text-foreground overflow-x-auto whitespace-nowrap">
                    TTx1Lr2ZLAKW7CBJMKyxfWXZoEFrjH8edN
                  </code>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => {
                      navigator.clipboard.writeText("TTx1Lr2ZLAKW7CBJMKyxfWXZoEFrjH8edN");
                      toast({
                        title: "Copied!",
                        description: "USDT TRC-20 address copied to clipboard",
                      });
                    }}
                    className="shrink-0 h-9 w-9 rounded-lg"
                  >
                    <Copy className="size-4" />
                  </Button>
                </div>
              </div>
            </div>

            <div className="text-center text-xs text-muted-foreground">
              Any amount helps us maintain and improve the platform. Thank you! ❤️
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Change Password Modal */}
      <Dialog open={isPasswordModalOpen} onOpenChange={setIsPasswordModalOpen}>
        <DialogContent className="sm:max-w-[425px] rounded-2xl" aria-describedby="password-description">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold tracking-tight flex items-center gap-2">
              <Lock className="size-5 text-primary" />
              Change Password
            </DialogTitle>
            <p id="password-description" className="text-sm text-muted-foreground">
              Update your account password for better security
            </p>
          </DialogHeader>
          <div className="grid gap-6 py-4">
            <div className="grid gap-2">
              <Label htmlFor="current" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Current Password</Label>
              <Input id="current" type="password" placeholder="••••••••" className="rounded-xl border-border/50 focus:ring-primary/20" />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="new" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">New Password</Label>
              <Input id="new" type="password" placeholder="••••••••" className="rounded-xl border-border/50 focus:ring-primary/20" />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="confirm" className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Confirm New Password</Label>
              <Input id="confirm" type="password" placeholder="••••••••" className="rounded-xl border-border/50 focus:ring-primary/20" />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="submit"
              className="w-full rounded-xl font-bold uppercase tracking-wider h-11"
              onClick={handleChangePassword}
            >
              Update Password
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

    </>
  );
}
