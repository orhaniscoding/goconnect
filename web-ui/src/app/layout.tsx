// Root layout — tüm sayfaların ebeveyni
export const metadata = {
  title: 'GoConnect',
  description: 'Secure virtual network & chat — orhaniscoding',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="tr">
      <body>{children}</body>
    </html>
  );
}
