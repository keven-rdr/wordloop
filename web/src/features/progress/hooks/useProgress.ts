import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getDashboardOptions, getHeatmapOptions } from '../../../api/generated/@tanstack/react-query.gen';

// O painel da pessoa (sequência, acertos, previsão, palavras). A tradução das palavras segue o idioma do app.
export function useDashboard() {
  const { i18n } = useTranslation();
  return useQuery(getDashboardOptions({ query: { locale: i18n.language === 'en-US' ? 'en-US' : 'pt-BR' } }));
}

// Os dias de estudo das últimas 12 semanas (o servidor usa o dia de estudo da pessoa).
export function useHeatmap() {
  return useQuery(getHeatmapOptions());
}
