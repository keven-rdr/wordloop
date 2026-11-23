import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { Outlet } from '@tanstack/react-router';
import { getMeOptions } from '../api/generated/@tanstack/react-query.gen';
import { BottomNav } from '../shared/ui/nav/BottomNav';
import { SideNav } from '../shared/ui/nav/SideNav';
import { desktopQuery } from '../shared/ui/tokens/breakpoints';
import { size } from '../shared/ui/tokens/color.stylex';
import { useMedia } from '../shared/ui/useMedia';
import { navItems } from './nav/registry';

const styles = stylex.create({
  wrap: { flexGrow: 1, display: 'flex', flexDirection: 'column' },
  // No desktop o conteúdo começa depois do menu lateral fixo.
  besideNav: { paddingInlineStart: size.sideNav },
  content: { flexGrow: 1, display: 'flex', flexDirection: 'column' },
});

// Moldura das telas de quem entrou: conteúdo e menu (inferior no mobile, lateral a partir de `bp.desktop`), com a
// foto do usuário. Os itens vêm do registro das features (`features/*/feature.ts`).
export function AppLayout() {
  const { data } = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  const desktop = useMedia(desktopQuery);
  const side = Boolean(data) && desktop;
  return (
    <div {...stylex.props(styles.wrap, side && styles.besideNav)}>
      {data && desktop && <SideNav items={navItems} name={data.display_name} pictureUrl={data.picture_url} />}
      <div {...stylex.props(styles.content)}>
        <Outlet />
      </div>
      {data && !desktop && <BottomNav items={navItems} name={data.display_name} pictureUrl={data.picture_url} />}
    </div>
  );
}
