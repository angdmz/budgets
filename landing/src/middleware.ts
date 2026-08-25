import { NextRequest, NextResponse } from 'next/server';
import { locales, defaultLocale } from '@/i18n/config';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const pathnameHasLocale = locales.some(
    locale => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}` 
  );
  if (pathnameHasLocale) return;

  const acceptLanguage = request.headers.get('accept-language') || '';
  const detected = locales.find(l => acceptLanguage.toLowerCase().includes(l)) || defaultLocale;

  request.nextUrl.pathname = `/${detected}${pathname}`;
  return NextResponse.redirect(request.nextUrl);
}

export const config = {
  matcher: ['/((?!_next|robots\\.txt|sitemap\\.xml|favicon\\.ico|.*\\..*).*)'],
};
