import { useQuery } from "@tanstack/react-query";

export interface User {
  id: string;
  nickname: string;
  money: number;
  bets: number;
  won_bets: number;
  settled_bets: number;
  avg_odds: number;
  topup: number;
  last_topup_at?: string | null;
  auth_provider?: string;
  created: string;
  updated: string;
}

// JWT token management
export const getAccessToken = () => localStorage.getItem('access_token');
export const setAccessToken = (token: string) => localStorage.setItem('access_token', token);
export const removeAccessToken = () => localStorage.removeItem('access_token');

export const getRefreshToken = () => {
  // Try to get from localStorage first, then from cookies
  const token = localStorage.getItem('refresh_token');
  if (token) return token;

  // If not in localStorage, try to get from cookies
  const cookies = document.cookie.split(';');
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=');
    if (name === 'refresh_token') {
      return decodeURIComponent(value);
    }
  }
  return null;
};

export const setRefreshToken = (token: string) => {
  localStorage.setItem('refresh_token', token);
};

export const removeRefreshToken = () => {
  localStorage.removeItem('refresh_token');
  // Also clear from cookies
  document.cookie = 'refresh_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
};

export function useUser() {
  const accessToken = getAccessToken();

  return useQuery({
    queryKey: ["user", accessToken], // Include token in queryKey for reactivity
    queryFn: async () => {
      try {
        if (!accessToken) {
          return null;
        }

        const res = await fetch("/api/auth/user", {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
          },
          credentials: 'include', // Still include for cookies refresh
        });

        if (!res.ok) {
          if (res.status === 401) {
            // Token expired or invalid, try to refresh
            const refreshToken = getRefreshToken();
            if (refreshToken) {
              try {
                const refreshRes = await fetch('/api/auth/refresh', {
                  method: 'POST',
                  credentials: 'include',
                  headers: {
                    'Content-Type': 'application/json',
                  },
                });

                if (refreshRes.ok) {
                  const refreshData = await refreshRes.json();
                  if (refreshData.access_token) {
                    setAccessToken(refreshData.access_token);
                    // Retry the original request with new token
                    const retryRes = await fetch("/api/auth/user", {
                      method: 'GET',
                      headers: {
                        'Authorization': `Bearer ${refreshData.access_token}`,
                        'Content-Type': 'application/json',
                      },
                      credentials: 'include',
                    });

                    if (retryRes.ok) {
                      const retryData = await retryRes.json();
                      if (retryData.success && retryData.user) {
                        return retryData.user as User;
                      }
                    }
                  }
                }
              } catch (refreshError) {
                console.debug("Token refresh failed:", refreshError);
              }
            }
            return null; // User not authenticated - this is normal
          }
          throw new Error("Failed to fetch user");
        }

        const data = await res.json();
        if (data.success && data.user) {
          return data.user as User;
        }
        return null;
      } catch (error) {
        // Silently handle network errors to avoid console spam
        console.debug("User authentication check failed:", error);
        return null;
      }
    },
    retry: false,
    staleTime: 5000,
    enabled: !!accessToken, // Only run query if token exists
  });
}
