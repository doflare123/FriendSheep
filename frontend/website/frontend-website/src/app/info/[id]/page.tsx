import { notFound } from 'next/navigation';
import ReactMarkdown from 'react-markdown';
import Footer from '@/components/Footer';
import styles from '@/styles/info/info.module.css';
import { promises as fs } from 'fs';
import path from 'path';

export default async function InfoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const filePath = path.join(process.cwd(), 'public', `${id}.md`);
    const markdownContent = await fs.readFile(filePath, 'utf8');

    return (
      <div className="bgPage">
        <div className={styles.pageContainer}>
          <div className={styles.contentWrapper}>
            <article className={styles.markdownContent}>
              <ReactMarkdown>{markdownContent}</ReactMarkdown>
            </article>
          </div>
        </div>
        
        <div className={styles.footerWrapper}>
          <Footer />
        </div>
      </div>
    );
  } catch (error) {
    notFound();
  }
}