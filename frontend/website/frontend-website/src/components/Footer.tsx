import SocialIcon from './SocialIcon';
import Image from 'next/image';

const Footer: React.FC = () => {
    return (
        <footer className="bg-white border-t-2 border-[#316BC2] py-8 px-6 font-inter text-xl">
            <div className="max-w-[120rem] mx-auto flex justify-between items-start md:flex-row flex-col md:space-y-0 space-y-6">
                {/* Левая часть */}
                <div className="flex flex-col space-y-2 -mt-6">
                    <div className="flex flex-col">
                        <Image 
                        src="/logo.png" 
                        alt="FriendShip" 
                        width={64} 
                        height={64} 
                        className="rounded-[1.0rem] ml-4" 
                        />
                        <h3 className="text-2xl font-semibold text-[#37A2E6]">FriendShip</h3>
                        <div className="w-150 h-0.5 bg-[#316BC2]"></div>
                    </div>
                    
                    <div className="footer-links">
                        <a href="/privacy" className="footer-link">Политика конфиденциальности</a>
                        <span className="text-gray-400">·</span>
                        <a href="/about" className="footer-link">О нас</a>
                        <span className="text-gray-400">·</span>
                        <a href="/feedback" className="footer-link">Обратная связь</a>
                        
                    </div>

                    <div className="footer-links">
                        <span className="text-gray-400">·</span>
                        <a href="/complaints" className="footer-link">Подача заявки на подтверждение личности</a>
                    </div>
                    
                    <div className="text-gray-500 mt-5">
                        ©2025 NecroDwarf
                    </div>
                </div>
                {/* Правая часть */}
                <div className="flex md:flex-row md:space-x-8 md:items-center mt-8">
                    <div className="flex flex-col items-center space-y-3 -mt-6 text-center">
                        <Image 
                        src="/social/qr_mob.png" 
                        alt="QR код" 
                        width={128} 
                        height={128} 
                        />
                        <p className="text-black">Мобильное приложение</p>
                    </div>
                    
                    <div className="flex flex-col space-y-3 items-center">
                        <p className="text-gray-600">Мы в соц сетях • Поддержать нас</p>
                        <div className="flex space-x-3">
                        <SocialIcon 
                            href="https://vk.com" 
                            iconName="vk" 
                            alt="ВКонтакте" 
                            size={52}
                        />
                        <SocialIcon 
                            href="https://t.me" 
                            iconName="tg" 
                            alt="Telegram" 
                            size={52}
                        />
                        <SocialIcon 
                            href="https://discord.com" 
                            iconName="ds" 
                            alt="Discord" 
                            size={52}
                        />
                        <SocialIcon 
                            href="#" 
                            iconName="bs" 
                            alt="Поддержать" 
                            size={52}
                        />
                        </div>
                    </div>
                </div>
            </div>
        </footer>
    );
};

export default Footer;