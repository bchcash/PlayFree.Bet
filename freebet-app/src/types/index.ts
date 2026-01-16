export interface Bet {
  bet_id: string;
  user_id: string;
  match_id: string;
  bet_type: 'home' | 'draw' | 'away';
  bet_amount: number;
  odds: number;
  potential_win: number;
  status: 'pending' | 'won' | 'lost';
  home_team: string;
  away_team: string;
  created_at: string;
  updated_at: string;
  commence_time?: string;
}

export interface Game {
  id: string;
  api_id: string;
  home_team: string;
  away_team: string;
  commence_time: string;
  home_odds: string;
  draw_odds: string;
  away_odds: string;
  completed: boolean;
  calculated: boolean;
  result: string | null;
  home_score: number | null;
  away_score: number | null;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  email: string;
  nickname: string;
  money: number;
  topup: number;
  last_topup_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Player {
  id: string;
  nickname: string;
  money: number;
  bets: number;
  wonBets: number;
  settledBets: number;
  avgOdds: number;
  topup: number;
  created_at: string;
  updated_at: string;
}
