import SocialIcon from './SocialIcon';
import Image from 'next/image';
import styles from '@/styles/Footer.module.css';
import QRCode from '@/components/QRCode';

const Footer: React.FC = () => {
    return (
        <footer className={styles.footer}>
            <div className={styles.footerContainer}>
                {/* Левая часть */}
                <div className={styles.leftSection}>
                    <div className={styles.logoSection}>
                        <Image 
                            src="/logo.png" 
                            alt="FriendShip" 
                            width={64} 
                            height={64} 
                            className={styles.logo}
                        />
                        <h3 className={styles.title}>FriendShip</h3>
                        <div className={styles.divider}></div>
                    </div>
                    
                    <div className={styles.footerLinks}>
                        <a href="/info/privacy" className={styles.footerLink} target="_blank" rel="noopener noreferrer">
                            Политика конфиденциальности
                        </a>
                        <span className={styles.separator}>·</span>
                        <a href="/info/about" className={styles.footerLink} target="_blank" rel="noopener noreferrer">
                            О нас
                        </a>
                        <span className={styles.separator}>·</span>
                        <a href="https://docs.google.com/forms/d/e/1FAIpQLScq8yseDrHN2dQ7eTfon6KqiohGzPAE95FRoyh8KkaFWuTB9Q/viewform?usp=dialog" className={styles.footerLink} target="_blank" rel="noopener noreferrer">
                            Обратная связь
                        </a>
                    </div>

                    <div className={styles.footerLinks}>
                        <span className={styles.separator}>·</span>
                        <a href="https://docs.google.com/forms/d/e/1FAIpQLSeFJ1U2cmAUS2ddxCYduNsA8cuzzuVLSQ3-3zm0JOls_9ZQSw/viewform?usp=publish-editor" className={styles.footerLink} target="_blank" rel="noopener noreferrer">
                            Подача заявки на подтверждение личности
                        </a>
                    </div>
                    
                    <div className={styles.copyright}>
                        ©2025 NecroDwarf
                    </div>
                </div>

                {/* Правая часть */}
                <div className={styles.rightSection}>
                    <div className={styles.qrSection}>
                        <QRCode size={128} showText={false} />
                        <p className={styles.qrText}>Мобильное приложение</p>
                    </div>
                    
                    <div className={styles.socialSection}>
                        <p className={styles.socialText}>Мы в соц сетях • Поддержать нас</p>
                        <div className={styles.socialIcons}>
                            <SocialIcon 
                                href="https://vk.com/club234284540" 
                                alt="ВКонтакте" 
                                size={52}
                            />
                            <SocialIcon 
                                href="https://t.me/necrodwarfs" 
                                alt="Telegram" 
                                size={52}
                            />
                            <SocialIcon 
                                href="https://discord.gg/V94MaRe4tk" 
                                alt="Discord" 
                                size={52}
                            />
                            <SocialIcon 
                                href="https://www.youtube.com/@CmapnepTV" 
                                alt="YouTube" 
                                size={52}
                            />
                            <SocialIcon 
                                href="https://www.donationalerts.com/r/necrodwarf" 
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