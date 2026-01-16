// Team logos are stored in /images/teams/
// Optimized WebP format with PNG fallback for better performance
export const teamLogos: Record<string, string> = {
  "Manchester City": "/images/teams/manchester-city.webp",
  "Arsenal": "/images/teams/arsenal-fc.webp",
  "Liverpool": "/images/teams/liverpool-fc.webp",
  "Aston Villa": "/images/teams/aston-villa.webp",
  "Tottenham Hotspur": "/images/teams/tottenham-hotspur.webp",
  "Chelsea": "/images/teams/chelsea-fc.webp",
  "Newcastle United": "/images/teams/newcastle-united.webp",
  "Manchester United": "/images/teams/manchester-united.webp",
  "West Ham United": "/images/teams/west-ham-united.webp",
  "Brighton and Hove Albion": "/images/teams/brighton-hove-albion.webp",
  "Wolverhampton Wanderers": "/images/teams/wolverhampton-wanderers.webp",
  "Fulham": "/images/teams/fulham-fc.webp",
  "Bournemouth": "/images/teams/bournemouth.webp",
  "Everton": "/images/teams/everton-fc.webp",
  "Brentford": "/images/teams/brentford.webp",
  "Nottingham Forest": "/images/teams/nottingham-forest.webp",
  "Crystal Palace": "/images/teams/crystal-palace.webp",
  "Leeds United": "/images/teams/leeds-united.webp",
  "Burnley": "/images/teams/burnley.webp",
  "Southampton": "/images/teams/southampton.webp"
};

// Fallback mapping for browsers that don't support WebP
export const getLogoWithFallback = (teamName: string): string => {
  const webpLogo = teamLogos[teamName];
  if (!webpLogo) return '';

  // Check WebP support
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = 1;
  const hasWebPSupport = canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0;

  if (hasWebPSupport) {
    return webpLogo;
  } else {
    // Fallback to PNG
    return webpLogo.replace('.webp', '.png');
  }
};
