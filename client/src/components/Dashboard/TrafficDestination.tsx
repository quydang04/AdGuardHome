import React from 'react';
import { useTranslation } from 'react-i18next';

import Card from '../ui/Card';

const COUNTRY_NAMES: Record<string, string> = {
    AF: 'Afghanistan', AL: 'Albania', DZ: 'Algeria', AD: 'Andorra', AO: 'Angola',
    AG: 'Antigua and Barbuda', AR: 'Argentina', AM: 'Armenia', AU: 'Australia', AT: 'Austria',
    AZ: 'Azerbaijan', BS: 'Bahamas', BH: 'Bahrain', BD: 'Bangladesh', BB: 'Barbados',
    BY: 'Belarus', BE: 'Belgium', BZ: 'Belize', BJ: 'Benin', BT: 'Bhutan',
    BO: 'Bolivia', BA: 'Bosnia and Herzegovina', BW: 'Botswana', BR: 'Brazil', BN: 'Brunei',
    BG: 'Bulgaria', BF: 'Burkina Faso', BI: 'Burundi', KH: 'Cambodia', CM: 'Cameroon',
    CA: 'Canada', CV: 'Cape Verde', CF: 'Central African Republic', TD: 'Chad', CL: 'Chile',
    CN: 'China', CO: 'Colombia', KM: 'Comoros', CG: 'Congo', CD: 'DR Congo',
    CR: 'Costa Rica', CI: "Côte d'Ivoire", HR: 'Croatia', CU: 'Cuba', CY: 'Cyprus',
    CZ: 'Czechia', DK: 'Denmark', DJ: 'Djibouti', DM: 'Dominica', DO: 'Dominican Republic',
    EC: 'Ecuador', EG: 'Egypt', SV: 'El Salvador', GQ: 'Equatorial Guinea', ER: 'Eritrea',
    EE: 'Estonia', SZ: 'Eswatini', ET: 'Ethiopia', FJ: 'Fiji', FI: 'Finland',
    FR: 'France', GA: 'Gabon', GM: 'Gambia', GE: 'Georgia', DE: 'Germany',
    GH: 'Ghana', GR: 'Greece', GD: 'Grenada', GT: 'Guatemala', GN: 'Guinea',
    GW: 'Guinea-Bissau', GY: 'Guyana', HT: 'Haiti', HN: 'Honduras', HU: 'Hungary',
    IS: 'Iceland', IN: 'India', ID: 'Indonesia', IR: 'Iran', IQ: 'Iraq',
    IE: 'Ireland', IL: 'Israel', IT: 'Italy', JM: 'Jamaica', JP: 'Japan',
    JO: 'Jordan', KZ: 'Kazakhstan', KE: 'Kenya', KI: 'Kiribati', KP: 'North Korea',
    KR: 'South Korea', KW: 'Kuwait', KG: 'Kyrgyzstan', LA: 'Laos', LV: 'Latvia',
    LB: 'Lebanon', LS: 'Lesotho', LR: 'Liberia', LY: 'Libya', LI: 'Liechtenstein',
    LT: 'Lithuania', LU: 'Luxembourg', MG: 'Madagascar', MW: 'Malawi', MY: 'Malaysia',
    MV: 'Maldives', ML: 'Mali', MT: 'Malta', MH: 'Marshall Islands', MR: 'Mauritania',
    MU: 'Mauritius', MX: 'Mexico', FM: 'Micronesia', MD: 'Moldova', MC: 'Monaco',
    MN: 'Mongolia', ME: 'Montenegro', MA: 'Morocco', MZ: 'Mozambique', MM: 'Myanmar',
    NA: 'Namibia', NR: 'Nauru', NP: 'Nepal', NL: 'Netherlands', NZ: 'New Zealand',
    NI: 'Nicaragua', NE: 'Niger', NG: 'Nigeria', MK: 'North Macedonia', NO: 'Norway',
    OM: 'Oman', PK: 'Pakistan', PW: 'Palau', PA: 'Panama', PG: 'Papua New Guinea',
    PY: 'Paraguay', PE: 'Peru', PH: 'Philippines', PL: 'Poland', PT: 'Portugal',
    QA: 'Qatar', RO: 'Romania', RU: 'Russia', RW: 'Rwanda', KN: 'Saint Kitts and Nevis',
    LC: 'Saint Lucia', VC: 'Saint Vincent', WS: 'Samoa', SM: 'San Marino',
    ST: 'São Tomé and Príncipe', SA: 'Saudi Arabia', SN: 'Senegal', RS: 'Serbia',
    SC: 'Seychelles', SL: 'Sierra Leone', SG: 'Singapore', SK: 'Slovakia', SI: 'Slovenia',
    SB: 'Solomon Islands', SO: 'Somalia', ZA: 'South Africa', SS: 'South Sudan',
    ES: 'Spain', LK: 'Sri Lanka', SD: 'Sudan', SR: 'Suriname', SE: 'Sweden',
    CH: 'Switzerland', SY: 'Syria', TW: 'Taiwan', TJ: 'Tajikistan', TZ: 'Tanzania',
    TH: 'Thailand', TL: 'Timor-Leste', TG: 'Togo', TO: 'Tonga', TT: 'Trinidad and Tobago',
    TN: 'Tunisia', TR: 'Turkey', TM: 'Turkmenistan', TV: 'Tuvalu', UG: 'Uganda',
    UA: 'Ukraine', AE: 'United Arab Emirates', GB: 'United Kingdom', US: 'United States',
    UY: 'Uruguay', UZ: 'Uzbekistan', VU: 'Vanuatu', VE: 'Venezuela', VN: 'Vietnam',
    YE: 'Yemen', ZM: 'Zambia', ZW: 'Zimbabwe',
};

const FLAG_EMOJI: Record<string, string> = {};
Object.keys(COUNTRY_NAMES).forEach((code) => {
    const codePoints = code.split('').map((c) => 0x1f1e6 + c.charCodeAt(0) - 65);
    FLAG_EMOJI[code] = String.fromCodePoint(...codePoints);
});

interface TopCountryEntry {
    name: string;
    count: number;
}

interface TrafficDestinationProps {
    topCountries: TopCountryEntry[];
    subtitle: string;
    refreshButton: React.ReactNode;
}

const getCountryColor = (percentage: number): string => {
    if (percentage >= 30) return '#1a6b3c';
    if (percentage >= 15) return '#2fb344';
    if (percentage >= 5) return '#5dd57a';
    if (percentage >= 1) return '#a3e4b8';
    return '#d4f4de';
};

const TrafficDestination = ({ topCountries, subtitle, refreshButton }: TrafficDestinationProps) => {
    const { t } = useTranslation();

    const totalQueries = topCountries.reduce((sum, entry) => sum + entry.count, 0);

    const countryData = topCountries.map((entry) => ({
        code: entry.name,
        name: COUNTRY_NAMES[entry.name] || entry.name,
        flag: FLAG_EMOJI[entry.name] || '',
        count: entry.count,
        percentage: totalQueries > 0 ? (entry.count / totalQueries) * 100 : 0,
    }));

    const hasData = countryData.length > 0;

    return (
        <Card
            title={t('traffic_destination')}
            subtitle={subtitle}
            bodyType="card-body"
            refresh={refreshButton}>
            <div className="traffic-destination">
                <div className="traffic-destination__desc">
                    {t('traffic_destination_desc')}
                </div>
                {!hasData ? (
                    <div className="traffic-destination__empty">
                        {t('traffic_destination_empty')}
                    </div>
                ) : (
                    <>
                        <div className="traffic-destination__bar">
                            {countryData.slice(0, 10).map((country) => (
                                <div
                                    key={country.code}
                                    className="traffic-destination__bar-segment"
                                    style={{
                                        width: `${Math.max(country.percentage, 2)}%`,
                                        backgroundColor: getCountryColor(country.percentage),
                                    }}
                                    title={`${country.flag} ${country.name}: ${country.count} (${country.percentage.toFixed(1)}%)`}
                                />
                            ))}
                        </div>

                        <div className="traffic-destination__list">
                            {countryData.slice(0, 10).map((country, index) => (
                                <div key={country.code} className="traffic-destination__item">
                                    <span className="traffic-destination__rank">{index + 1}</span>
                                    <span
                                        className="traffic-destination__color"
                                        style={{ backgroundColor: getCountryColor(country.percentage) }}
                                    />
                                    <span className="traffic-destination__flag">{country.flag}</span>
                                    <span className="traffic-destination__name">{country.name}</span>
                                    <span className="traffic-destination__count">
                                        {country.count.toLocaleString()}
                                    </span>
                                    <span className="traffic-destination__pct">
                                        {country.percentage.toFixed(1)}%
                                    </span>
                                </div>
                            ))}
                        </div>
                    </>
                )}
            </div>
        </Card>
    );
};

export default TrafficDestination;
