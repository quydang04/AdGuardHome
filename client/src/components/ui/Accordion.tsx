import React, { useState, ReactNode } from 'react';
import './Accordion.css';
import '../Settings/Clients/Service.css';

export interface AccordionItemProps {
    id: string;
    title: string;
    children: ReactNode;
    defaultOpen?: boolean;
    className?: string;
    disabled?: boolean;
}

type Props = AccordionItemProps & {
  isOpen: boolean;
  onToggle: () => void;
};

const AccordionItem = (props: Props) => {
    const {
        id,
        title,
        children,
        isOpen,
        onToggle,
        disabled,
        className = '',
    } = props;
    return (
        <section className={`accordion-item ${className}`} data-testid={`accordion-item-${id}`}>
            <header className="accordion-item__header">
                <div className="accordion-item__toggle-wrapper">
                    <button
                        type="button"
                        className={`accordion-item__toggle ${isOpen ? 'accordion-item__toggle--open' : ''}`}
                        onClick={() => { if (!disabled) { onToggle(); } }}
                        aria-expanded={isOpen}
                        aria-controls={`accordion-content-${id}`}
                        aria-disabled={disabled}
                        disabled={disabled}
                    >
                        <span className="accordion-item__icon" aria-hidden="true">
                            <svg width="24" height="24" viewBox="0 0 24 24">
                                <use xlinkHref="#chevron-down" />
                            </svg>
                        </span>
                        <h3 className="accordion-item__title">{title}</h3>
                    </button>
                </div>
            </header>

            <div
                id={`accordion-content-${id}`}
                className={`accordion-item__content ${isOpen ? 'accordion-item__content--open' : ''}`}
                aria-hidden={!isOpen}
            >
                <div className="accordion-item__content-inner">
                    {children}
                </div>
            </div>
        </section>
    );
};

type AccordionProps = {
    items: any[];
    allowMultiple?: boolean;
    className?: string;
};

export const Accordion: React.FC<AccordionProps> = ({
    items,
    allowMultiple = false,
    className = '',
}) => {
    const [openItems, setOpenItems] = useState<Set<string>>(() => {
        const defaultOpen = new Set<string>();
        items.forEach((item) => {
            if (item.defaultOpen) {
                defaultOpen.add(item.id);
            }
        });
        return defaultOpen;
    });

    const toggleItem = (itemId: string) => {
        setOpenItems((prev) => {
            const newOpenItems = new Set(prev);

            if (newOpenItems.has(itemId)) {
                newOpenItems.delete(itemId);
            } else {
                if (!allowMultiple) {
                    newOpenItems.clear();
                }
                newOpenItems.add(itemId);
            }

            return newOpenItems;
        });
    };

    return (
        <div className={`accordion ${className}`}>
            {items.map((item) => (
                <AccordionItem
                    key={item.id}
                    id={item.id}
                    title={item.title}
                    isOpen={openItems.has(item.id)}
                    onToggle={() => toggleItem(item.id)}
                    disabled={item.disabled}
                >
                    {item.children}
                </AccordionItem>
            ))}
        </div>
    );
};

export default Accordion;