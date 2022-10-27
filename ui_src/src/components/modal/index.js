// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useEffect } from 'react';
import { Modal } from 'antd';

import Button from '../button';

const TransitionsModal = ({
    height,
    width,
    rBtnText,
    lBtnText,
    rBtnDisabled,
    lBtnDisabled,
    header,
    isLoading,
    open = false,
    displayButtons = true,
    lBtnClick,
    clickOutside,
    rBtnClick,
    children,
    hr = false,
    className
}) => {
    const contentStyle = {
        height: height,
        overflowY: 'auto',
        borderTop: hr ? '1px solid #EAEAEA' : null
    };

    return (
        <Modal
            wrapClassName={className || 'modal-wrapper'}
            title={header}
            open={open}
            width={width}
            onCancel={() => clickOutside()}
            bodyStyle={contentStyle}
            centered
            destroyOnClose={true}
            footer={
                displayButtons
                    ? [
                          <div key="left" className="btnContainer">
                              <button className="cancel-button" disabled={lBtnDisabled} onClick={() => lBtnClick()}>
                                  {lBtnText}
                              </button>
                              <Button
                                  className="modal-btn"
                                  width="83px"
                                  height="32px"
                                  placeholder={rBtnText}
                                  disabled={rBtnDisabled}
                                  colorType="white"
                                  radiusType="circle"
                                  backgroundColorType={'purple'}
                                  fontSize="12px"
                                  fontWeight="600"
                                  isLoading={isLoading}
                                  onClick={() => {
                                      rBtnClick();
                                  }}
                              />
                          </div>
                      ]
                    : null
            }
        >
            {children}
        </Modal>
    );
};

export default TransitionsModal;
