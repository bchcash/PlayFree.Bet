import { useState, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';

interface OptimizedImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src: string;
  alt: string;
  fallbackSrc?: string;
  className?: string;
  priority?: boolean;
  onLoad?: () => void;
  onError?: () => void;
}

export function OptimizedImage({
  src,
  alt,
  fallbackSrc,
  className,
  priority = false,
  onLoad,
  onError,
  ...props
}: OptimizedImageProps) {
  const [isLoading, setIsLoading] = useState(true);
  const [hasError, setHasError] = useState(false);
  const [currentSrc, setCurrentSrc] = useState(src);
  const imgRef = useRef<HTMLImageElement>(null);

  // WebP support detection
  const supportsWebP = () => {
    const canvas = document.createElement('canvas');
    canvas.width = canvas.height = 1;
    return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0;
  };

  // Handle WebP fallback
  useEffect(() => {
    if (!supportsWebP() && src.endsWith('.webp') && fallbackSrc) {
      setCurrentSrc(fallbackSrc);
    }
  }, [src, fallbackSrc]);

  const handleLoad = () => {
    setIsLoading(false);
    onLoad?.();
  };

  const handleError = () => {
    setHasError(true);
    setIsLoading(false);

    // Try fallback if available and not already using it
    if (fallbackSrc && currentSrc !== fallbackSrc) {
      setCurrentSrc(fallbackSrc);
      setHasError(false);
      setIsLoading(true);
    } else {
      onError?.();
    }
  };

  return (
    <div className="relative">
      {isLoading && (
        <div className={cn(
          "absolute inset-0 bg-muted animate-pulse rounded",
          className
        )} />
      )}

      {!hasError && (
        <img
          ref={imgRef}
          src={currentSrc}
          alt={alt}
          loading={priority ? "eager" : "lazy"}
          decoding="async"
          className={cn(
            "transition-opacity duration-300",
            isLoading ? "opacity-0" : "opacity-100",
            className
          )}
          onLoad={handleLoad}
          onError={handleError}
          {...props}
        />
      )}

      {hasError && fallbackSrc && (
        <div className={cn(
          "bg-muted rounded flex items-center justify-center text-muted-foreground text-xs",
          className
        )}>
          {alt.charAt(0).toUpperCase()}
        </div>
      )}
    </div>
  );
}