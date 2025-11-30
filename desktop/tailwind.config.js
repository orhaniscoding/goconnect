/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                // GoConnect dark theme
                'gc-dark': {
                    900: '#202225',
                    800: '#2f3136',
                    700: '#36393f',
                    600: '#40444b',
                    500: '#4f545c',
                },
                'gc-primary': '#5865f2',
                'gc-green': '#3ba55c',
                'gc-red': '#ed4245',
                'gc-yellow': '#faa61a',
            },
        },
    },
    plugins: [],
}
