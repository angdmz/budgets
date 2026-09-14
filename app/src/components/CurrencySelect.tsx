import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import type { CurrencyInfo } from '../lib/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

interface CurrencySelectProps {
  value: string;
  onChange: (currency: string) => void;
}

export default function CurrencySelect({ value, onChange }: CurrencySelectProps) {
  const { data: currencies } = useQuery({
    queryKey: ['currencies'],
    queryFn: async () => {
      const response = await axios.get<CurrencyInfo[]>(`${API_BASE_URL}/currencies`);
      return response.data;
    },
    staleTime: Infinity,
  });

  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      aria-label="Currency"
      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
      data-testid="currency-select"
    >
      {(currencies ?? []).map((currency) => (
        <option key={currency.code} value={currency.code}>
          {currency.code}
        </option>
      ))}
    </select>
  );
}
