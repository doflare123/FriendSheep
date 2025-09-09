import ProfilePage from './ProfilePage';

interface ProfilePageProps {
  params: {
    id: string;
  };
}

export default function Page({ params }: ProfilePageProps) {
  return <ProfilePage params={params} />;
}