import { Helmet } from 'react-helmet-async';
import { useLocation } from 'wouter';

interface HelmetManagerProps {
  title?: string;
  description?: string;
  keywords?: string;
  image?: string;
  url?: string;
}

export function HelmetManager({
  title,
  description,
  keywords,
  image,
  url
}: HelmetManagerProps) {
  const [location] = useLocation();

  // Базовые настройки для сайта
  const baseTitle = "FreeBet Guru - Virtual Betting Platform";
  const baseDescription = "FreeBet Guru - Virtual betting platform with live odds, player statistics, and leaderboard. Place virtual bets on football matches.";
  const baseKeywords = "virtual betting, free bets, football betting, odds, bookmaker, sports betting, online gambling, freebet, bet, betting guru, bet on sports, gambling online, gambling guru, premier league, epl betting, epl odds";
  const baseImage = "https://freebet.guru/favicon.png";
  const baseUrl = `https://freebet.guru${location}`;

  // Определяем контент в зависимости от маршрута
  const getPageMeta = () => {
    const path = location;

    if (path === '/') {
      return {
        title: "FreeBet Guru - Virtual Football Betting",
        description: "Experience the thrill of virtual football betting! Place free bets on Premier League matches, track live odds, and compete with other players on our leaderboard.",
        keywords: `${baseKeywords}, virtual football betting, premier league betting, free football bets, soccer betting`,
        image: baseImage,
        url: baseUrl
      };
    }

    if (path === '/leaderboard') {
      return {
        title: "Leaderboard - Top Players | FreeBet Guru",
        description: "Check out the top virtual bettors on FreeBet Guru. See player rankings, win rates, profit/loss statistics, and betting performance metrics.",
        keywords: `${baseKeywords}, betting leaderboard, top bettors, player rankings, betting statistics, profit loss`,
        image: baseImage,
        url: baseUrl
      };
    }

    if (path.startsWith('/player/')) {
      const playerName = path.split('/player/')[1]?.replace('-', ' ') || 'Player';
      return {
        title: `${playerName} - Player Profile | FreeBet Guru`,
        description: `View ${playerName}'s betting history, statistics, win rate, and performance on FreeBet Guru. Track player progress and betting patterns.`,
        keywords: `${baseKeywords}, player profile, betting history, player statistics, win rate, betting performance`,
        image: baseImage,
        url: baseUrl
      };
    }

    if (path === '/dashboard') {
      return {
        title: "Dashboard - My Bets | FreeBet Guru",
        description: "Access your personal betting dashboard on FreeBet Guru. View your betting history, balance, profit/loss, and manage your virtual betting account.",
        keywords: `${baseKeywords}, betting dashboard, my bets, account balance, profit loss, betting history`,
        image: baseImage,
        url: baseUrl
      };
    }

    // Для неизвестных маршрутов возвращаем базовые мета-теги
    return {
      title: baseTitle,
      description: baseDescription,
      keywords: baseKeywords,
      image: baseImage,
      url: baseUrl
    };
  };

  const pageMeta = getPageMeta();

  // Используем переданные пропсы или значения по умолчанию для страницы
  const finalTitle = title || pageMeta.title;
  const finalDescription = description || pageMeta.description;
  const finalKeywords = keywords || pageMeta.keywords;
  const finalImage = image || pageMeta.image;
  const finalUrl = url || pageMeta.url;

  return (
    <Helmet>
      {/* Basic meta tags */}
      <title>{finalTitle}</title>
      <meta name="description" content={finalDescription} />
      <meta name="keywords" content={finalKeywords} />

      {/* Canonical URL */}
      <link rel="canonical" href={finalUrl} />

      {/* Open Graph meta tags */}
      <meta property="og:title" content={finalTitle} />
      <meta property="og:description" content={finalDescription} />
      <meta property="og:type" content="website" />
      <meta property="og:url" content={finalUrl} />
      <meta property="og:image" content={finalImage} />
      <meta property="og:image:width" content="512" />
      <meta property="og:image:height" content="512" />
      <meta property="og:image:alt" content={finalTitle} />
      <meta property="og:site_name" content="FreeBet Guru" />
      <meta property="og:locale" content="en_US" />

      {/* Twitter Card meta tags */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:title" content={finalTitle} />
      <meta name="twitter:description" content={finalDescription} />
      <meta name="twitter:image" content={finalImage} />
      <meta name="twitter:image:alt" content={finalTitle} />
      <meta name="twitter:site" content="@freebet_guru" />
      <meta name="twitter:creator" content="@freebet_guru" />

      {/* Additional SEO meta tags */}
      <meta name="robots" content="index, follow, max-snippet:-1, max-image-preview:large, max-video-preview:-1" />
      <meta name="language" content="English" />
      <meta name="author" content="FreeBet Guru" />

      {/* Structured Data - Organization */}
      <script type="application/ld+json">
        {JSON.stringify({
          "@context": "https://schema.org",
          "@type": "Organization",
          "name": "FreeBet Guru",
          "url": "https://freebet.guru",
          "logo": "https://freebet.guru/favicon.png",
          "description": "Virtual betting platform with live odds and player statistics",
          "sameAs": [
            "https://t.me/freebet_guru",
            "https://x.com/freebet_guru"
          ]
        })}
      </script>

      {/* Structured Data - WebSite */}
      <script type="application/ld+json">
        {JSON.stringify({
          "@context": "https://schema.org",
          "@type": "WebSite",
          "name": "FreeBet Guru",
          "url": "https://freebet.guru",
          "description": "Virtual betting platform for football matches",
          "potentialAction": {
            "@type": "SearchAction",
            "target": "https://freebet.guru/search?q={search_term_string}",
            "query-input": "required name=search_term_string"
          }
        })}
      </script>
    </Helmet>
  );
}