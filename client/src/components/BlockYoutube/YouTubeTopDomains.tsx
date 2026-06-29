import React from 'react';
import { useTranslation } from 'react-i18next';

import { YoutubeTopDomain } from '../../initialState';

type Props = {
    domains: YoutubeTopDomain[];
    title: string;
};

const YouTubeTopDomains = ({ domains, title }: Props) => {
    const { t } = useTranslation();
    const maxCount = domains.length > 0 ? domains[0].count : 1;

    return (
        <div className="yt-top-domains">
            <h6 className="yt-top-domains__title">{title}</h6>
            {domains.length === 0 ? (
                <p className="yt-no-data">{t('youtube_no_stats')}</p>
            ) : (
                <table className="yt-top-domains__table">
                    <tbody>
                        {domains.map((d) => (
                            <tr key={d.domain}>
                                <td className="yt-top-domains__domain">{d.domain}</td>
                                <td className="yt-top-domains__count">{d.count.toLocaleString()}</td>
                                <td className="yt-top-domains__bar-cell">
                                    <div
                                        className="yt-top-domains__bar"
                                        style={{ width: `${(d.count / maxCount) * 100}%` }}
                                    />
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </div>
    );
};

export default YouTubeTopDomains;
