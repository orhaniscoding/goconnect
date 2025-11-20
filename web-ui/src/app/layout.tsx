// App Router root layout
import { NotificationProvider } from '../contexts/NotificationContext'
import ToastContainer from '../components/ToastContainer'

export const metadata = {
  title: 'GoConnect',
  description: 'Secure virtual network & chat — orhaniscoding',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <NotificationProvider>
          {children}
          <ToastContainer />
        </NotificationProvider>
      </body>
    </html>
  )
}
