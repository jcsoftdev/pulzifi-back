export function PlusIcon({
  className = '',
}: Readonly<{
  className?: string
}>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="11"
      height="11"
      viewBox="0 0 11 11"
      fill="none"
      className={className}
      role="img"
    >
      <title>Plus</title>
      <path
        d="M3.12674 5.21095H7.29516M5.21095 3.12674V7.29516M1.56359 0.521484H8.85832C9.43386 0.521484 9.90042 0.98805 9.90042 1.56359V8.85832C9.90042 9.43386 9.43386 9.90042 8.85832 9.90042H1.56359C0.98805 9.90042 0.521484 9.43386 0.521484 8.85832V1.56359C0.521484 0.98805 0.98805 0.521484 1.56359 0.521484Z"
        stroke="currentColor"
        strokeOpacity="0.88"
        strokeWidth="1.0421"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}
