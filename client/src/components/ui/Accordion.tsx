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

export interface AccordionProps {
    items: AccordionItemProps[];
    allowMultiple?: boolean;
    className?: string;
    onGroupToggle?: (groupId: string, enabled: boolean) => void;
    groupEnabledStates?: Record<string, boolean>;
}

const AccordionItem: React.FC<AccordionItemProps & { 
    isOpen: boolean;
    onToggle: () => void;
    onGroupToggle?: (groupId: string, enabled: boolean) => void;
    groupEnabled?: boolean;
}> = ({ id, title, children, isOpen, onToggle, onGroupToggle, groupEnabled = true, className = '', disabled = false }) => {
    return (
        <section className={`accordion-item ${className}`} data-testid={`accordion-item-${id}`}>
            <header className="accordion-item__header">
                <div className="accordion-item__toggle-wrapper">
                    <button
                        type="button"
                        className={`accordion-item__toggle ${isOpen ? 'accordion-item__toggle--open' : ''}`}
                        onClick={onToggle}
                        aria-expanded={isOpen}
                        aria-controls={`accordion-content-${id}`}
                    >
                        <span className="accordion-item__icon" aria-hidden="true">
                            <svg width="24" height="24" viewBox="0 0 24 24">
                                <use xlinkHref="#chevron-down" />
                            </svg>
                        </span>
                        <h3 className="accordion-item__title">{title}</h3>
                    </button>

                    {onGroupToggle && (
                        <label className="accordion-item__group-switch">
                            <input
                                type="checkbox"
                                checked={groupEnabled}
                                onChange={(e) => onGroupToggle(id, e.target.checked)}
                                className="custom-switch-input"
                                disabled={disabled}
                            />
                            <span className="service__switch custom-switch-indicator"></span>
                        </label>
                    )}
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

export const Accordion: React.FC<AccordionProps> = ({ 
    items,
    allowMultiple = false,
    className = '',
    onGroupToggle,
    groupEnabledStates = {}
    }) => {
        const [openItems, setOpenItems] = useState<Set<string>>(() => {
        const defaultOpen = new Set<string>();
        items.forEach(item => {
            if (item.defaultOpen) {
                defaultOpen.add(item.id);
            }
        });
        return defaultOpen;
    });

    const toggleItem = (itemId: string) => {
        setOpenItems(prev => {
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
            {items.map(item => (
                <AccordionItem
                key={item.id}
                {...item}
                isOpen={openItems.has(item.id)}
                onToggle={() => toggleItem(item.id)}
                onGroupToggle={onGroupToggle}
                groupEnabled={groupEnabledStates[item.id] ?? true}
                />
            ))}
        </div>
    );
};

export default Accordion;