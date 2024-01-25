// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';
import React from 'react';
import { Drawer } from 'antd';

/**
 * MSDrawer Component
 *
 * A custom wrapper around the Ant Design `Drawer` component with additional features.
 *
 * @param {string} title - The title displayed at the top of the drawer.
 * @param {string} placement - The placement of the drawer. Can be one of 'top', 'right', 'bottom', or 'left'.
 * @param {string} size - The size of the drawer. Can be one of 'small', 'medium', or 'large'.
 * @param {string} width - The width of the drawer. Fox example '600px'.
 * @param {string} height - The height of the drawer. Fox example '700px'.
 * @param {string} className - The class name for the drawer. Fox example 'custom-drawer'.
 * @param {callback} onClose - A callback function called when the drawer is closed.
 * @param {boolean} destroyOnClose - Whether to destroy the drawer content when it's closed.
 * @param {boolean} open - Whether the drawer is open or closed.
 * @param {object} maskStyle - Additional CSS styles for the overlay mask when the drawer is open.
 * @param {object} headerStyle - Additional CSS styles for the header style of drawer.
 * @param {object} bodyStyle - Additional CSS styles for the body style of drawer.
 * @param {ReactNode} closeIcon - Custom icon or element to use as the close button.
 * @param {ReactNode} children - The content to be displayed inside the drawer.
 *
 * @returns {ReactNode} - A React component representing the custom drawer.
 */

const MSDrawer = ({ title, placement, size, width, height, className, onClose, destroyOnClose, open, maskStyle, headerStyle, bodyStyle, closeIcon, children, mask }) => {
    return (
        <Drawer
            title={title}
            placement={placement}
            size={size}
            width={width}
            height={height}
            className={className}
            onClose={onClose}
            destroyOnClose={destroyOnClose}
            open={open}
            maskStyle={maskStyle}
            headerStyle={headerStyle}
            bodyStyle={bodyStyle}
            closeIcon={closeIcon}
            mask={mask}
        >
            {children}
        </Drawer>
    );
};

export default MSDrawer;
