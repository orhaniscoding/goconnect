// App Router root layout
export const metadata = {
  title: 'GoConnect',
  description: 'Secure virtual network & chat — orhaniscoding',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>{children}</body>
    </html>
  )
}
