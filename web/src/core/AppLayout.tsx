import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { Outlet } from '@tanstack/react-router';
import { getMeOptions } from '../api/generated/@tanstack/react-query.gen';
import { BottomNav } from '../shared/ui/nav/BottomNav';

const styles = stylex.create({
  wrap: { flexGrow: 1, display: 'flex', flexDirection: 'column' },
  content: { flexGrow: 1, display: 'flex', flexDirection: 'column' },
});

// Moldura das telas de quem entrou: conteudo e menu inferior (com a foto do usuario).
export function AppLayout() {
  const { data } = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  return (
    <div {...stylex.props(styles.wrap)}>
      <div {...stylex.props(styles.content)}>
        <Outlet />
      </div>
      {data && <BottomNav name={data.display_name} pictureUrl={data.picture_url} />}
    </div>
  );
}
